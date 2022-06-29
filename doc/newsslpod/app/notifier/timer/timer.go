// Package timer provides ...
package timer

import (
	"fmt"
	"time"

	"mysslee_qcloud/app/notifier/db"
	"mysslee_qcloud/app/notifier/prom"
	"mysslee_qcloud/model"

	"github.com/sirupsen/logrus"
)

var duration = time.Second * 2

// Start task
func Start() {
	msgs, err := db.GetNoticeMsgs(1, 50)
	if err != nil {
		logrus.Info("monitorNotice: ", err)
	}
	if len(msgs) == 0 {
		time.AfterFunc(duration, Start)
	} else {
		for _, msg := range msgs {
			go retrySendNotice(msg)
		}
		Start()
	}
}

// MaxRetryCount the retry count if any error
const MaxRetryCount = 3

// 重试机制发送
func retrySendNotice(msg *model.NoticeMsg) {
	prom.PromNoticeStatus.WithLabelValues("total").Inc()

	var tryCount int
RETRY:
	_, err := sendMsg2SubAccountToQCloud(msg)
	if err != nil {
		tryCount++
		if tryCount < MaxRetryCount {
			time.Sleep(time.Second)
			goto RETRY
		} else {
			msg.NoticedResult = fmt.Sprintf("%s -> {%v}", "Max tryCount 3", err)
		}
	} else {
		msg.NoticedResult = "success"
	}

	// 通知状态
	err = db.UpNoticeMsg(msg.Id, map[string]interface{}{
		"noticed_result": msg.NoticedResult,
	})
	if err != nil {
		prom.PromNoticeStatus.WithLabelValues("failed").Inc()
		logrus.Error("retrySendNotice.UpNoticeMsg: ", err)
	}
}
