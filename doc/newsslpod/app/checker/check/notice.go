// Package check provides ...
package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"mysslee_qcloud/app/checker/db"
	"mysslee_qcloud/app/checker/redis"
	"mysslee_qcloud/config"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"
	redis2 "mysslee_qcloud/redis"

	"github.com/sirupsen/logrus"
)

type GroupNotice struct {
	Domain      string    // 域名
	Port        string    // 端口
	NoticeType  string    // 通知类型
	NoticeObj   string    // 通知目标
	FiringCount int       // 等待时间内
	CreatedAt   time.Time // 创建时间
}

// WarnNotice 获取通知用户，校验通知间隔
func WarnNotice(result *model.DomainResult) {
	subs, err := db.GetUsersWatchDomain(result.Id)
	if err != nil {
		logrus.Error("WarnNotice.GetUsersWatchDomain: ", err)
		return
	}

	for _, sub := range subs {
		now := time.Now()
		key := fmt.Sprintf("%d:%s:%s:%s", sub.AccountId, result.Domain, result.Port, result.IP)
		// 是否需要通知
		if !sub.NoticedAt.IsZero() || redis.ShouldWarnNotice(key) {
			logrus.Info("should not WarnNotice: ",
				!sub.NoticedAt.IsZero(), redis.ShouldWarnNotice(key))
			continue
		}
		sub.NoticedAt = now

		err = redis.SetWarnNotice(key, now.Unix())
		if err != nil {
			logrus.Error("WarnNotice.SetWarnNotice: ", err)
		}
		// 更新
		err = db.UpdateAccountDomain(sub.Id, map[string]interface{}{
			"noticed_at": now,
		})
		if err != nil {
			logrus.Error("WarnNotice.UpdateAccountDomain: ", err)
			continue
		}
		// 用户信息
		a, err := db.GetAccountById(sub.AccountId)
		if err != nil {
			logrus.Error("sendNotice.GetAccountById: ", err)
			return
		}
		// 开关信息
		nInfo, err := db.GetNoticeInfo(a.Uin)
		if err != nil {
			logrus.Error("sendNotice.GetNoticeInfo: ", err)
			return
		}
		// 通知消息
		resp, err := qcloud.GetAccountInfoFromQCloud(a.Uin)
		if err != nil {
			logrus.Error("sendNotice.GetAccountInfoFromQCloud: ", err)
			return
		}
		content := fmt.Sprintf(`{"Nickname":"%s","Domain":"%s","Port":"%s","Time":"%s","TrustStatus":"%s"}`,
			resp.GetString("nickname"),
			result.Domain,
			result.Port,
			result.LastFastDetectionTime.Local().Format("2006-01-02 15:04:05"),
			result.TrustStatus)
		msg := &model.NoticeMsg{
			Uin:      nInfo.Uin,
			Type:     nInfo.NoticeType,
			Msg:      content,
			Language: "zh",
			// NoticeAlready: model.NoticeNone,
			CreatedAt: now,
		}
		sendNotice(a.Uin, msg, model.NoticeTypeCertStatus)
	}
}

// sendNotice 保存通知消息到数据库
func sendNotice(uin string, msg *model.NoticeMsg, ntype int) {
	// 获取额度信息
	plan, err := getCalculatedLimit(uin)
	if err != nil {
		logrus.Error("sendNotice.GetCalculatedLimit: ", err)
		return
	}
	var result string
	// if msg.Type&model.NoticeEmail == model.NoticeEmail {
	ok, err := db.IncrLimitConsume(uin, time.Now().Format("2006-01"), plan.MaxAllowEmailWarnCount)
	if err != nil {
		logrus.Error("sendNotice.IncrLimitConsume: ", err)
		return
	}
	if !ok {
		result = "limit Exceeded"
	}
	// }
	// if msg.Type&model.NoticePhoneSMS == model.NoticePhoneSMS {
	// 	ok, err := redis.GetNoticeLimit(uin, model.PhoneLimitName, plan.MaxAllowEmailWarnCount)
	// 	if err != nil {
	// 		logrus.Error("sendNotice.GetNoticeLimit.PhoneSMS: ", err)
	// 		return
	// 	}
	// 	if !ok {
	// 		result += "phone: out of max count 、"
	// 		msg.NoticeAlready |= model.NoticePhoneSMS
	// 	}
	// }
	// if msg.Type&model.NoticeWechat == model.NoticeWechat {
	// 	ok, err := redis.GetNoticeLimit(uin, model.WechatLimitName, plan.MaxAllowEmailWarnCount)
	// 	if err != nil {
	// 		logrus.Error("sendNotice.GetNoticeLimit.Wechat: ", err)
	// 		return
	// 	}
	// 	if !ok {
	// 		result += "wechat: out of max count "
	// 		msg.NoticeAlready |= model.NoticeWechat
	// 	}
	// }
	if result != "" {
		msg.NoticedAt = msg.CreatedAt
		msg.NoticedResult = result
	}
	err = db.AddNoticeMsg(msg)
	if err != nil {
		logrus.Error("sendNotice.AddNoticeMsg: ", err)
	}
}

// 获取限制
func getCalculatedLimit(uin string) (*model.PlanInfo, error) {
	plan := new(model.PlanInfo)
	data, err := redis.GetPlanLimit(uin)
	if err != nil {
		if err != redis2.Nil {
			return nil, err
		}
		// get calculated limit
		// request backend
		ips, err := redis2.ScanApp(redis2.BackendApp)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			req, err := http.NewRequest(http.MethodGet,
				fmt.Sprintf("http://%s:%d/api/plan/%s", ip, config.Conf.Backend.Listen, uin),
				nil)
			if err != nil {
				logrus.Error("getCalculatedLimit.NewRequest: ", err)
				continue
			}
			req.SetBasicAuth("qcloud", "bab658e3c1176a664ccbdd74a4f4b9a5")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logrus.Error("getCalculatedLimit.DefaultClient: ", err)
				continue
			}
			defer resp.Body.Close()

			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Error("getCalculatedLimit.ReadAll: ", err)
				continue
			}
			if resp.StatusCode != 200 {
				err = errors.New(string(data))
				continue
			}
			break
		}
	}
	return plan, json.Unmarshal(data, plan)
}
