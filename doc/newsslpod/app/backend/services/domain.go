// Package services provides ...
package services

import (
	"context"
	"encoding/json"
	"errors"
	"mysslee_qcloud/app/backend/aggregation"
	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/common"
	"mysslee_qcloud/core/myconn"
	"mysslee_qcloud/dns"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"
	"mysslee_qcloud/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
)

func init() {
	qcloud.Register("CreateDomain", HandleDomainAdd)
	qcloud.Register("DeleteDomain", HandleDomainDel)
	qcloud.Register("DescribeDomains", HandleDomainSearch)
	qcloud.Register("ChangeDomainSwitch", HandleDomainNoticeSwitch)
	qcloud.Register("RefreshDomain", HandleDomainRefresh)
	qcloud.Register("ResolveDomain", HandleDomainDNSResolve)
	qcloud.Register("DescribeDomainTags", HandleDomainTags)
	qcloud.Register("ModifyDomainTags", HandleChangeTags)
	qcloud.Register("DescribeDomainCerts", HandleDomainCertDetail)
	// 下面俩接口暂时不用
	qcloud.Register("DescribeDomainRegionalDetail", HandleDescribeDomainRegionalDetail)
	qcloud.Register("DescribeDomainIPDetail", HandleDescribeDomainIPDetail)
	// 域名监控详情接口（带多ip多证书）
	qcloud.Register("DescribeDomainCertsDetail", HandleDescribeDomainCertsDetail)
	qcloud.Register("CreateDomainMonitor", HandleCreateDomainMonitor)
	qcloud.Register("ModifyDomainIPs", HandleModifyDomainIPs)
	qcloud.Register("DescribeDomainIPs", HandleDescribeDomainIPs)
}

// DomainInfo 域名信息
type DomainInfo struct {
	Domain     string
	IP         string
	Port       string
	ServerType int
	Tags       []string
	Notice     bool
}

// CoreAddDomain 添加域名
func CoreAddDomain(req qcloud.RequestBack) string {
	a := req.GetValue("account").(*model.Account)

	serverType := req.GetInt("ServerType")
	domain := req.GetString("Domain")
	port := req.GetString("Port")
	ip := req.GetString("IP")
	notice := req.GetBool("Notice")
	tags := req.GetString("Tags")
	info := &DomainInfo{
		Domain:     strings.ToLower(strings.TrimSpace(domain)),
		Port:       strings.TrimSpace(port),
		IP:         strings.TrimSpace(ip),
		Notice:     notice,
		ServerType: serverType,
	}
	if serverType == 4 && info.Port == "" {

	}
	if tags != "" {
		info.Tags = strings.Split(tags, ",")
	}
	code, _, _, _ := handleAddDomains([]*DomainInfo{info}, a, true)
	return code
}

// HandleCreateDomainMonitor 创建域名监控-新版多IP
func HandleCreateDomainMonitor(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	a := req.GetValue("account").(*model.Account)

	serverType := req.GetInt("ServerType")
	domain := req.GetString("Domain")
	port := req.GetString("Port")
	notice := req.GetBool("Notice")
	tags := req.GetString("Tags")
	if req.GetValue("IPPorts") == nil {
		req.SetKeys("IPPorts", []interface{}{})
	}
	IPPortsInput := req.GetValue("IPPorts").([]interface{})
	isAutoDetect := req.GetBool("IsAutoDetect")
	info := &DomainInfo{
		Domain:     strings.ToLower(strings.TrimSpace(domain)),
		Port:       strings.TrimSpace(port),
		IP:         "",
		Notice:     notice,
		ServerType: serverType,
	}
	if serverType == 4 && info.Port == "" {

	}
	if tags != "" {
		info.Tags = strings.Split(tags, ",")
	}
	code, _, _, domainIds := handleAddDomains([]*DomainInfo{info}, a, true)
	if len(domainIds) <= 0 {
		logrus.Error("HandleCreateDomainMonitor.handleAddDomains: ", code)
		return nil, code
	}
	IPPorts := []*model.IPPort{}
	for _, i := range IPPortsInput {
		IPPort := model.IPPort{}
		if _, ok := i.(map[string]interface{})["IP"]; ok {
			IPPort.IP = i.(map[string]interface{})["IP"].(string)
		} else {
			return nil, qcloud.ErrInvalidParameter
		}
		if _, ok := i.(map[string]interface{})["Port"]; ok {
			IPPort.Port = i.(map[string]interface{})["Port"].(string)
		} else {
			IPPort.Port = "443"
		}
		IPPorts = append(IPPorts, &IPPort)
	}
	IPPortsJSON, err := json.Marshal(IPPorts)
	if err != nil {
		logrus.Error("HandleCreateDomainMonitor.jsonMarshal: ", err)
		return nil, qcloud.ErrFailedOperation
	}
	domainIPs := &model.DomainIps{
		DomainID:     domainIds[0],
		Uin:          a.Uin,
		IpPorts:      string(IPPortsJSON),
		IsAutoDetect: isAutoDetect,
	}
	err = db.SaveDomainIPs(domainIPs)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}

// HandleModifyDomainIPs 修改域名绑定的IP
func HandleModifyDomainIPs(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	a := req.GetValue("account").(*model.Account)
	domainID := req.GetInt("DomainId")
	if req.GetValue("IPPorts") == nil {
		req.SetKeys("IPPorts", []interface{}{})
	}
	IPPortsInput := req.GetValue("IPPorts").([]interface{})
	isAutoDetect := req.GetBool("IsAutoDetect")
	IPPorts := []*model.IPPort{}
	for _, i := range IPPortsInput {
		IPPort := model.IPPort{}
		if _, ok := i.(map[string]interface{})["IP"]; ok {
			IPPort.IP = i.(map[string]interface{})["IP"].(string)
		} else {
			return nil, qcloud.ErrInvalidParameter
		}
		if _, ok := i.(map[string]interface{})["Port"]; ok {
			IPPort.Port = i.(map[string]interface{})["Port"].(string)
		} else {
			IPPort.Port = "443"
		}
		IPPorts = append(IPPorts, &IPPort)
	}
	IPPortsJSON, err := json.Marshal(IPPorts)
	if err != nil {
		logrus.Error("HandleCreateDomainMonitor.jsonMarshal: ", err)
		return nil, qcloud.ErrFailedOperation
	}
	domainIPs := &model.DomainIps{
		DomainID:     domainID,
		Uin:          a.Uin,
		IpPorts:      string(IPPortsJSON),
		IsAutoDetect: isAutoDetect,
	}
	err = db.SaveDomainIPs(domainIPs)
	if err != nil {
		logrus.Error("HandleModifyDomainIPs.SaveDomainIPs: ", err)
		return nil, qcloud.ErrFailedOperation
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}

// HandleDomainAdd 添加域名
// @Summary 添加监控域名
// @Description 通过域名端口添加监控, 0 web, 1 smtp, 2 imap, 3 pops
// @Tags 监控管理
// @Accept json
// @Produce json
// @Param action formData string true "CreateDomain"
// @Param serviceType formData string true "sslpod"
// @Param ServerType formData int true "监控的服务器类型" Enums(0,1,2,3)
// @Param Domain formData string true "添加的域名"
// @Param Port formData string true "添加的端口"
// @Param IP formData string false "指定域名的IP，可选"
// @Param Notice formData bool false "是否开启通知告警"
// @Param Tags formData string false "给域名添加标签，多个以逗号隔开"
// @Success 200 {object} qcloud.ResponseBack
// @Router /domains [post]
func HandleDomainAdd(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains").Inc()

	errCode := CoreAddDomain(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

func getDefaultPort(serverType int) (string, error) {
	var port string
	switch myconn.ServerType(serverType) {
	case myconn.Web:
		port = "443"
	case myconn.SMTP:
		port = "465"
	case myconn.IMAP:
		port = "993"
	case myconn.POP3:
		port = "995"
	default:
		return "", errors.New("Invalid ServerType")
	}
	return port, nil
}

func handleAddDomains(infos []*DomainInfo, a *model.Account, only bool) (code string, success int, failed int,
	domainIds []int) {
	defer func() {
		if success > 0 {
			// 设置聚合flag
			err := db.SetAccountAggrFlagOnly(a.Id)
			if err != nil {
				logrus.Error("fullCheck.SetAccountAggrFlagOnly: ", err)
			}
		}
	}()
	// 检查参数
	domainResults, noticeCount, code := checkDomainInfo(infos)
	if code != "" {
		return
	}
	// 判断限制
	total, err := db.CountAttentionDomain(a.Id)
	if err != nil {
		logrus.Error("HandleAddDomain.CountAttentionDomain ", err)
		code = qcloud.ErrInternalError
		return
	}
	plan, err := db.GetCalculatedLimit(a.Uin)
	if err != nil {
		logrus.Error("HandleAddDomain.GetCalculatedLimit ", err)
		code = qcloud.ErrInternalError
		return
	}
	if total+len(infos) > plan.MaxAllowAddDomainCount {
		code = qcloud.ErrLimitedAddDomain
		return
	}
	if noticeCount > 0 {
		count, err := db.CountNoticeDomainNumber(a.Id)
		if err != nil {
			logrus.Error("HandleAddDomain.CountNoticeDomainNumber ", err)
			code = qcloud.ErrInternalError
			return
		}
		if count+noticeCount > plan.MaxAllowMonitoringCount {
			code = qcloud.ErrLimitedMonitorDomain
			return
		}
	}
	// 添加到数据库
	domainIds = []int{}
	for i, domainResult := range domainResults {
		domainId, exist := db.IsExistDomainResultReturnDomainId(domainResult.Domain,
			domainResult.IP, domainResult.Port, domainResult.ServerType)
		if !exist {
			data, _ := json.Marshal(model.CalculateStatusForUnknown())
			domainResult.DomainStatus = string(data)
			err = db.InsertDomainResultWithModel(domainResult)
			if err != nil {
				logrus.Error("HandleAddDomain.InsertDomainResultWithModel ", err)
				code = qcloud.ErrInternalError
				return
			}
			domainId = domainResult.Id
		}
		domainIds = append(domainIds, domainId)
		// 是否已存在关系
		have, err := db.UserHaveAttentionDomain(a.Id, domainId)
		if err != nil {
			logrus.Error("HandleAddADomain.UserHaveAttentionDomain ", err)
			code = qcloud.ErrInternalError
			return
		}
		if !have {
			accountDomain := &model.AccountDomain{
				AccountId: a.Id,
				DomainId:  domainId,
				Notice:    infos[i].Notice,
				CreatedAt: time.Now().UTC(),
			}
			err = db.InsertAccountDomainRelation(accountDomain, infos[i].Tags)
			if err != nil {
				logrus.Error("HandleAddADomain.InsertAccountDomainRelation ", err)
				code = qcloud.ErrInternalError
				return
			}
			success++
			// 添加成功后，如果该域名信息已经禁用，改为启用状态
			err = db.UpdateLoseEfficacy(domainId, false)
			if err != nil {
				logrus.Error("HandleAddADomain.UpdateLoseEfficacy ", err)
				code = qcloud.ErrInternalError
				return
			}
		} else if only {
			code = qcloud.ErrRepetitionAdd
			return
		} else {
			failed++
		}
	}
	return
}

func checkDomainInfo(infos []*DomainInfo) ([]*model.DomainResult, int, string) {
	var (
		domainResults []*model.DomainResult
		noticeCount   int
	)
	for _, info := range infos {
		if info.ServerType < 0 || info.ServerType > 4 {
			return nil, 0, qcloud.ErrInvalidServerType
		} else if info.Port == "" {
			var err error
			info.Port, err = getDefaultPort(info.ServerType)
			if err != nil {
				return nil, 0, qcloud.ErrInvalidServerType
			}
		}
		if !utils.ValidatePort(info.Port) {
			return nil, 0, qcloud.ErrInvalidPort
		}
		punycodeDomain, err := idna.ToASCII(info.Domain)
		if err != nil {
			logrus.Error("checkDomainInfo.ToASCII ", err)
			return nil, 0, qcloud.ErrInvalidDomain
		}
		if !utils.ValidateDomain2(punycodeDomain) && !utils.ValidateIP(punycodeDomain) {
			return nil, 0, qcloud.ErrInvalidDomain
		}
		if utils.ValidateIP(punycodeDomain) {
			info.IP = punycodeDomain
		}
		domainFlag := 0
		if info.IP != "" {
			if !utils.ValidateIP(info.IP) {
				return nil, 0, qcloud.ErrInvalidIP
			}
			domainFlag |= model.DomainFlagBindIP
		}
		for _, tag := range info.Tags {
			if !utils.ValidateTagName(tag) {
				return nil, 0, qcloud.ErrInvalidTagName
			}
		}
		if len(info.Tags) > 3 {
			return nil, 0, qcloud.ErrTooManyTag
		}
		if info.Notice {
			noticeCount++
		}
		info.Domain, err = idna.ToUnicode(info.Domain)
		if err != nil {
			logrus.Error("checkDomainInfo.ToUnicode ", err)
			return nil, 0, qcloud.ErrInvalidDomain
		}
		domainResult := &model.DomainResult{
			Domain:              info.Domain,
			Port:                info.Port,
			IP:                  info.IP,
			ServerType:          info.ServerType,
			PunyCodeDomain:      punycodeDomain,
			FullDetectionResult: []byte("{}"),
			DomainFlag:          domainFlag,
			Grade:               "unknown",
		}
		domainResults = append(domainResults, domainResult)
	}
	return domainResults, noticeCount, ""
}

// CoreDomainDel 删除域名
func CoreDomainDel(req qcloud.RequestBack) string {
	a := req.GetValue("account").(*model.Account)

	domainId := req.GetInt("DomainId")
	if domainId == 0 {
		return qcloud.ErrInvalidParameterValue
	}
	exist, err := db.UserHaveAttentionDomain(a.Id, domainId)
	if err != nil {
		logrus.Error("HandleAddADomain.UserHaveAttentionDomain ", err)
		return qcloud.ErrInternalError
	}
	if !exist {
		return qcloud.ErrUnauthorizedOperation
	}
	// 删除
	err = db.DeleteRelation(a.Id, domainId)
	if err != nil {
		logrus.Error("CoreDomainDel.DeleteRelation: ", err)
		return qcloud.ErrInternalError
	}
	// 删除成功后计算
	count, err := db.CountAttentionAccountAmount(domainId)
	if err != nil {
		logrus.Error("CoreDomainDel.CountAttentionAccountAmount: ", err)
		return qcloud.ErrInternalError
	}
	// 关注该域名的人数为0，把改域名设置成失效状态
	if count == 0 {
		db.UpdateLoseEfficacy(domainId, true)
	}

	// 发送重新聚合请求
	agg := &aggregation.AggregateType{
		AccountId:  a.Id,
		FromDomain: false,
	}
	aggregation.AggrHandler.SendAggregateRequest(agg)

	return ""
}

// HandleDomainDel 删除域名
// @Summary 删除监控的域名
// @Description 通过域名ID删除
// @Tags 监控管理
// @Accept json
// @Produce json
// @Param action query string true "DeleteDomain"
// @Param serviceType query string true "sslpod"
// @Param DomainId path string true "域名列表中的ID"
// @Success 200 {object} qcloud.ResponseBack
// @Router /domains/{domainId} [delete]
func HandleDomainDel(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/:domainId").Inc()

	errCode := CoreDomainDel(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

// CoreSearch 搜索
func CoreSearch(req qcloud.RequestBack) (*Paging, string) {
	a := req.GetValue("account").(*model.Account)
	total, err := db.CountAttentionDomain(a.Id)
	if err != nil {
		logrus.Error("CoreSearch.CountAttentionDomain: ", err)
		return nil, qcloud.ErrInternalError
	}
	maxAllow, currentMonitor, maxAllowAdd, err := getAccountAllows(a)
	if err != nil {
		logrus.Error("CoreSearch.getAccountAllows: ", err)
		return nil, qcloud.ErrInternalError
	}

	var (
		paging = &Paging{
			Total:                  total,
			AllowMonitoringCount:   maxAllow,
			CurrentMonitoringCount: currentMonitor,
			AllowMaxAddDomain:      maxAllowAdd,
		}
		code string
	)

	searchType := req.GetString("SearchType")
	switch searchType {
	case common.SearchNone:
		code = handleFindAttentionDomains(req, paging, a.Id)
	case common.SearchTags:
		code = handleSearchByTags(req, paging, a.Id)
	case common.SearchSecureGrade:
		code = handleSearchBySecureLevel(req, paging, a.Id)
	case common.SearchBrand:
		code = handleSearchByBrand(req, paging, a.Id)
	case common.SearchCode:
		code = handleSearchByCode(req, paging, a.Id)
	case common.SearchHash:
		code = handleSearchByHash(req, paging, a.Id)
	case common.SearchDomain:
		code = handleSearchByDomain(req, paging, a.Id)
	case common.SearchLimit:
		// return paging
	default:
		return nil, qcloud.ErrInvalidSearchType
	}
	return paging, code
}

// HandleDomainSearch 搜索域名
// @Summary 搜索域名，域名列表
// @Description 通过searchType搜索已经添加的域名, 这里逻辑比较复杂
// @Tags 监控管理
// @Accept json
// @Produce json
// @Param action query string true "DescribeDomains"
// @Param serviceType query string true "sslpod"
// @Param Offset query int true "偏移量"
// @Param Limit query int true "获取数量"
// @Param SearchType query string true "搜索的类型" Enums(none,tags,grade,brand,code,hash,limit)
// @Param Tag query string false "tag"
// @Param Grade query string false "等级"
// @Param Brand query string false "品牌名"
// @Param Code query string false "混合搜索"
// @Param Hash query string false "证书hash"
// @Success 200 {object} services.Paging
// @Router /domains/search [post]
func HandleDomainSearch(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/search").Inc()

	data, errCode := CoreSearch(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"Data":      data,
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

// 处理分页信息
func handlePageInfo(req qcloud.RequestBack) (int, int) {
	offset := req.GetInt("Offset")
	if offset < 0 {
		offset = 0
	}
	limit := req.GetInt("Limit")
	if limit < 1 {
		limit = 10
	}
	return offset, limit

}

// 查看用户所关注的域名情况
func handleFindAttentionDomains(req qcloud.RequestBack, paging *Paging, accountId int) string {
	offset, limit := handlePageInfo(req)
	// 获取监控的域名
	results, total, err := db.GetAccountDomainResult(accountId, offset, limit)
	if err != nil {
		logrus.Error("CoreSearch.handleFindAttentionDomains: ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = total

	// 前端显示
	var siteInfos []*model.SiteInfo
	for _, result := range results {
		siteInfo := model.ChangeDomainResultToSiteInfo(result)
		if result.Children != nil {
			// 子信息
			for _, v := range result.Children {
				siteInfo.Children = append(siteInfo.Children, model.ChangeDomainResultToSiteInfo(v))
			}
		}
		siteInfos = append(siteInfos, siteInfo)
	}
	paging.Result = siteInfos

	return ""
}

func handleSearchByTags(req qcloud.RequestBack, paging *Paging, accountId int) string {
	tags := strings.Split(req.GetString("Tag"), ",")
	if len(tags) == 0 {
		return qcloud.ErrInvalidParameterValue
	}
	offset, limit := handlePageInfo(req)

	// searchTotal, err := db.CountAccountTagsWithArray(accountId, tags)
	// if err != nil {
	// 	logrus.Error("CoreSearch.CountAccountTagsWithArray: ", err)
	// 	return qcloud.ErrInternalError
	// }
	// paging.SearchTotal = searchTotal
	//
	// domainIds, err := db.SearchAccountTagsWithArray(accountId, tags, offset, limit)
	// if err != nil {
	// 	logrus.Error("CoreSearch.SearchAccountTagsWithArray: ", err)
	// 	return qcloud.ErrInternalError
	// }
	domainIds, total, err := db.SearchDomainIdsOfAccountByTags(accountId, tags, offset, limit)
	if err != nil {
		logrus.Error("CoreSearch.SearchDomainIdsOfAccountByTags: ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = total

	err = getSiteInfos(accountId, domainIds, paging)
	if err != nil {
		logrus.Error("CoreSearch.getSiteInfos: ", err)
		return qcloud.ErrInternalError
	}

	return ""
}

// 根据安全等级查询
func handleSearchBySecureLevel(req qcloud.RequestBack, paging *Paging, accountId int) string {
	// 验证安全等级是否正确
	secureLevel := req.GetString("Grade")
	secureLevel = strings.TrimSpace(secureLevel)
	if !common.IsCorrectSecureCode(secureLevel) {
		return qcloud.ErrInvalidParameterValue
	}
	if secureLevel == common.Unknown {
		secureLevel = common.SecureLevelUnknown
	}
	if secureLevel == "Ap" {
		secureLevel = "A+"
	}

	offset, limit := handlePageInfo(req)

	searchTotal, err := db.CountSearchBySecureGrade(accountId, secureLevel)
	if err != nil {
		logrus.Error("CoreSearch.CountSearchBySecureGrade: ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = searchTotal

	domainsIds, err := db.SearchBySecureGrade(accountId, secureLevel, offset, limit)
	if err != nil {
		logrus.Error("CoreSearch.SearchBySecureGrade: ", err)
		return qcloud.ErrInternalError
	}

	err = getSiteInfos(accountId, domainsIds, paging)
	if err != nil {
		logrus.Error("CoreSearch.getSiteInfos: ", err)
		return qcloud.ErrInternalError
	}
	return ""
}

// 根据证书品牌查询
func handleSearchByBrand(req qcloud.RequestBack, paging *Paging, accountId int) string {
	brand := strings.TrimSpace(req.GetString("Brand"))
	if brand == "" {
		return qcloud.ErrInvalidParameterValue
	}

	offset, limit := handlePageInfo(req)

	searchTotal, err := db.CountDomainByBrand(accountId, brand)
	if err != nil {
		logrus.Error("CoreSearch.CountDomainByBrand: ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = searchTotal

	domainIds, err := db.GetDomainByBrand(accountId, brand, offset, limit)
	if err != nil {
		logrus.Error("CoreSearch.GetDomainByBrand: ", err)
		return qcloud.ErrInternalError
	}
	err = getSiteInfos(accountId, domainIds, paging)
	if err != nil {
		logrus.Error("CoreSearch.getSiteInfos: ", err)
		return qcloud.ErrInternalError
	}

	return ""
}

// 根据Code码查询
// 进行混合查询
func handleSearchByCode(req qcloud.RequestBack, paging *Paging, accountId int) string {
	itemType := req.GetString("Item")
	status := req.GetString("Status")

	if strings.TrimSpace(itemType) == "" || strings.TrimSpace(status) == "" {
		return qcloud.ErrInvalidParameterValue
	}

	shift := model.GetShiftFromItemType(itemType)
	code := model.GetCodeFromShift(shift, status)

	if !model.IsCurrentShift(shift) {
		return qcloud.ErrInvalidParameterValue
	}
	offset, limit := handlePageInfo(req)

	domainIds, searchTotal, err := handleDomainsByCode(accountId, code, shift, offset, limit)
	if err != nil {
		logrus.Error("CoreSearch.handleSearchByCode ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = searchTotal

	err = getSiteInfos(accountId, domainIds, paging)
	if err != nil {
		logrus.Error("CoreSearch.getSiteInfos: ", err)
		return qcloud.ErrInternalError
	}

	return ""
}

// 通过证书hash查询
func handleSearchByHash(req qcloud.RequestBack, paging *Paging, accountId int) string {
	hash := req.GetString("Hash")
	if hash == "" {
		return qcloud.ErrInvalidParameterValue
	}

	offset, limit := handlePageInfo(req)

	domainIds, total, err := db.GetDomainByHash(accountId, hash, offset, limit)
	if err != nil {
		logrus.Error("CoreSearch.GetDomainByHash: ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = total
	err = getSiteInfos(accountId, domainIds, paging)
	if err != nil {
		logrus.Error("CoreSearch.getSiteInfos: ", err)
		return qcloud.ErrInternalError
	}
	return ""
}

func handleDomainsByCode(accountId int, code int64, shift uint, offset, limit int) (domainIds []int, searchTotal int,
	err error) {
	infos, err := db.GetDomainResultStatusAndId(accountId)
	if err != nil {
		return nil, 0, err
	}

	for _, info := range infos {
		status := &model.DomainStatus{}
		if info.Status == "{}" {
			continue
		}
		err := json.Unmarshal([]byte(info.Status), &status)
		if err != nil {
			continue
		}
		if model.VerifyCode(status.Status, code, shift) {
			domainIds = append(domainIds, info.DomainId)
		}

	}

	length := len(domainIds)
	if offset >= length {
		return nil, 0, errors.New("无法获取到需要的数据")
	}

	if offset+limit >= length {
		return domainIds[offset:], length, nil
	}
	return domainIds[offset : offset+limit], length, nil
}

// handleSearchByDomain 通过域名查询
func handleSearchByDomain(req qcloud.RequestBack, paging *Paging, accountId int) string {
	domain := req.GetString("Domain")
	if domain == "" {
		return qcloud.ErrInvalidDomain
	}
	punycodeDomain, err := idna.ToASCII(domain)
	if err != nil {
		logrus.Error("CoreSearch.ToASCII: ", err)
		return qcloud.ErrInvalidDomain
	}

	//offset, limit := handlePageInfo(req)

	domainIds, err := db.GetByDomain(accountId, punycodeDomain)
	if err != nil {
		logrus.Error("CoreSearch.GetByDomain: ", err)
		return qcloud.ErrInternalError
	}
	paging.SearchTotal = len(domainIds)
	err = getSiteInfos(accountId, domainIds, paging)
	if err != nil {
		logrus.Error("CoreSearch.getSiteInfos: ", err)
		return qcloud.ErrInternalError
	}
	return ""
}

func getNoticeLimit(account *model.Account, plan *model.PlanInfo) (maxAllow, currentMonitor int, err error) {
	maxAllow, err = getMaxMonitoringCount(account, plan)
	if err != nil {
		return 0, 0, err
	}
	currentMonitor, err = db.CountNoticeDomainNumber(account.Id)
	if err != nil {
		return 0, 0, err
	}
	return maxAllow, currentMonitor, nil

}

func getAccountAllows(account *model.Account) (maxAllow, currentMonitor, maxAllowAdd int, err error) {
	plan, err := db.GetCalculatedLimit(account.Uin)
	if err != nil {
		return 0, 0, 0, err
	}
	maxAllow, currentMonitor, err = getNoticeLimit(account, plan)
	if err != nil {
		return 0, 0, 0, err
	}
	maxAllowAdd = plan.MaxAllowAddDomainCount
	return maxAllow, currentMonitor, maxAllowAdd, nil
}

func getSiteInfos(accountId int, domainIds []int, paging *Paging) (err error) {
	var siteInfos []*model.SiteInfo
	for _, domainId := range domainIds {
		result, err := db.GetDomainResultWithOtherInfo(accountId, domainId)
		if err != nil {
			return err
		}

		siteInfos = append(siteInfos, model.ChangeDomainResultToSiteInfo(result))
	}
	paging.Result = siteInfos
	return nil
}

// 获取用户允许的最大监控数量
func getMaxMonitoringCount(a *model.Account, plan *model.PlanInfo) (int, error) {
	totalCount := plan.MaxAllowMonitoringCount
	// 忽略邀请注册的额度增加
	// invitationCount, err := db.CountSuccessInvitation(a.Id)
	// if err != nil {
	// 	if !gorm.IsRecordNotFoundError(err) {
	// 		return 0, err
	// 	}
	// }
	// if invitationCount >= 5 {
	// 	invitationCount = 5
	// }
	// totalCount += invitationCount * common.InvitationAdd
	// if a.EnterpriseId > 0 && plan.Id == model.ProductPlanBasic {
	// 	un := a.Email
	// 	if un == "" {
	// 		un = a.Phone.String
	// 	}
	// 	ep, err := db.GetEnterprise(un, a.EnterpriseId)
	// 	if err != nil {
	// 		logrus.Error(err)
	// 		return 0, err
	// 	}
	// 	if ep.Status == model.EnterpriseStatusVerified {
	// 		totalCount += 9
	// 	}
	// }
	return totalCount, nil
}

// CoreNoticeSwitch 开关通知
func CoreNoticeSwitch(req qcloud.RequestBack) string {
	a := req.GetValue("account").(*model.Account)

	domainId := req.GetInt("DomainId")
	if domainId < 1 {
		return qcloud.ErrInvalidParameterValue
	}
	// 是否是通知，判断限制
	isNotice, err := db.GetNotice(a.Id, domainId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return qcloud.ErrUnauthorizedOperation
		}
		logrus.Error("CoreNoticeSwitch.GetNotice: ", err)
		return qcloud.ErrInternalError
	}
	if !isNotice { // 需要开启
		plan, err := db.GetCalculatedLimit(a.Uin)
		if err != nil {
			logrus.Error("CoreNoticeSwitch.GetCalculatedLimit: ", err)
			return qcloud.ErrInternalError
		}
		// 计算最大量
		max, err := getMaxMonitoringCount(a, plan)
		if err != nil {
			logrus.Error("CoreNoticeSwitch.getMaxMonitoringCount: ", err)
			return qcloud.ErrInternalError
		}

		count, err := db.CountNoticeDomainNumber(a.Id)
		if err != nil {
			logrus.Error("CoreNoticeSwitch.CountNoticeDomainNumber: ", err)
			return qcloud.ErrInternalError
		}
		if count+1 > max {
			return qcloud.ErrLimitedMonitorDomain
		}
	}

	err = db.SetNotice(a.Id, domainId, isNotice)
	if err != nil {
		logrus.Error("CoreNoticeSwitch.SetNotice: ", err)
		return qcloud.ErrInternalError
	}

	return ""
}

// HandleDomainNoticeSwitch 通知开关
// @Summary 域名通知开关
// @Description 域名是否需要通知
// @Tags 监控管理
// @Accept json
// @Produce json
// @Param action formData string true "ChangeDomainSwitch"
// @Param serviceType formData string true "sslpod"
// @Param DomainId path string true "域名列表中的ID"
// @Success 200 {object} qcloud.ResponseBack
// @Router /domains/notice/{domainId} [put]
func HandleDomainNoticeSwitch(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/notice/:domainId").Inc()

	errCode := CoreNoticeSwitch(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

// CoreRefreshDomain 刷新域名
func CoreRefreshDomain(req qcloud.RequestBack) string {
	a := req.GetValue("account").(*model.Account)

	domainId := req.GetInt("DomainId")
	if domainId < 1 {
		return qcloud.ErrInvalidParameterValue
	}
	have, err := db.UserHaveAttentionDomain(a.Id, domainId)
	if err != nil {
		logrus.Error("HandleAddADomain.UserHaveAttentionDomain ", err)
		return qcloud.ErrInternalError
	}
	if !have {
		return qcloud.ErrAuthFailure
	}
	err = db.ResetDomainDetectionTime(domainId, time.Now())
	if err != nil {
		logrus.Error("CoreRefreshDomain.ResetDomainDetectionTime: ", err)
		return qcloud.ErrInternalError
	}
	return ""
}

// HandleDomainRefresh 强制刷新域名
// @Summary 刷新域名
// @Description 强制刷新域名检测
// @Tags 监控管理
// @Accept mpfd
// @Produce json
// @Param action query string true "RefreshDomain"
// @Param serviceType query string true "sslpod"
// @Param DomainId path string true "域名列表中的ID"
// @Success 200 {object} qcloud.ResponseBack
// @Router /domains/refresh/{domainId} [put]
func HandleDomainRefresh(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/refresh/:domainId").Inc()

	errCode := CoreRefreshDomain(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

// CoreDNSResolve 域名解析
func CoreDNSResolve(req qcloud.RequestBack) ([]string, string) {
	domain := req.GetString("Domain")
	if domain == "" {
		return nil, qcloud.ErrInvalidDomain
	}
	punycodeDomain, err := idna.ToASCII(domain)
	if err != nil {
		logrus.Error("CoreDNSResolve.ToASCII: ", err)
		return nil, qcloud.ErrInvalidDomain
	}
	if !utils.ValidateDomain2(punycodeDomain) {
		return nil, qcloud.ErrFailedResolveDomain
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ips, err := dns.LookupHost(ctx, punycodeDomain)
	if err != nil {
		logrus.Error("CoreDNSResolve.LookupHost: ", err)
		return nil, qcloud.ErrFailedResolveDomain
	}
	return ips, ""
}

// HandleDomainDNSResolve 解析域名
// @Summary 解析域名，获取IP
// @Description 解析域名获取IP地址
// @Tags 监控管理
// @Accept mpfd
// @Produce json
// @Param action query string true "ResolveDomain"
// @Param serviceType query string true "sslpod"
// @Param Domain query string true "解析的域名"
// @Success 200 {array} string "msg.Data"
// @Router /domains/resolve [get]
func HandleDomainDNSResolve(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/resolve").Inc()

	data, errCode := CoreDNSResolve(req)
	if errCode != "" {
		data = []string{}
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
			"Data":      data,
		},
	}, ""
}

// CoreAccountTags 获取账号标签
func CoreAccountTags(req qcloud.RequestBack) ([]string, string) {
	a := req.GetValue("account").(*model.Account)
	tags, err := db.GetAccountTags(a.Id)
	if err != nil {
		logrus.Error("CoreAccountTags.GetAccountTags: ", err)
		return nil, qcloud.ErrInternalError
	}

	return tags, ""
}

// HandleDomainTags 获取账号的tags
// @Summary tag列表
// @Description 获取所有tag
// @Tags 监控管理
// @Accept mpfd
// @Produce json
// @Param action query string true "DescribeDomainTags"
// @Param serviceType query string true "sslpod"
// @Success 200 {array} string
// @Router /domains/tags [get]
func HandleDomainTags(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/tags").Inc()

	data, errCode := CoreAccountTags(req)
	if errCode != "" {
		return nil, errCode
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
			"Data":      data,
		},
	}, ""
}

// CoreChangeTags 更新账号的tags
func CoreChangeTags(req qcloud.RequestBack) string {
	a := req.GetValue("account").(*model.Account)

	domainAccountId := req.GetInt("AccountDomainId")
	if domainAccountId < 1 {
		return qcloud.ErrInvalidParameterValue
	}

	var tags []string
	tagStr := req.GetString("Tags")
	if tagStr != "" {
		tags = strings.Split(tagStr, ",")
		if len(tags) > 3 {
			return qcloud.ErrTooManyTag
		}
	}
	for _, v := range tags {
		if !utils.ValidateTagName(v) {
			return qcloud.ErrInvalidTagName
		}
	}

	ad, err := db.GetAccountDomainWithAccount(a.Id, domainAccountId)
	if err != nil {
		logrus.Error("CoreChangeTags.GetAccountDomainWithAccount: ", err)
		return qcloud.ErrUnauthorizedOperation
	}

	err = db.DelDomainAccountTags(domainAccountId, ad.Tags)
	if err != nil {
		logrus.Error("CoreChangeTags.DelDomainAccountTags: ", err)
		return qcloud.ErrInternalError
	}

	err = db.AddDomainAccountTags(domainAccountId, tags)
	if err != nil {
		logrus.Error("CoreChangeTags.AddDomainAccountTags: ", err)
		return qcloud.ErrInternalError
	}

	return ""
}

// HandleChangeTags 给域名改变tag
// @Summary 改变域名tag
// @Description 改变域名tag
// @Tags 监控管理
// @Accept mpfd
// @Produce json
// @Param action query string true "ModifyDomainTags"
// @Param serviceType query string true "sslpod"
// @Param AccountDomainId path int true "域名ID"
// @Param Tags formData string true "更新后的tag，多个以逗号隔开"
// @Success 200 {object} qcloud.ResponseBack
// @Router /domains/tags/{AccountDomainId} [put]
func HandleChangeTags(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/tags/:id").Inc()

	errCode := CoreChangeTags(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

// CoreCertDetail 获取证书详情
func CoreCertDetail(req qcloud.RequestBack) ([]*model.CertInfoShow, string) {
	a := req.GetValue("account").(*model.Account)

	domainId := req.GetInt("DomainId")
	if domainId < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// get data
	infos, err := db.GetDomainCertDetail(a.Id, domainId)
	if err != nil {
		logrus.Error("CoreCertDetail.GetDomainCertDetail: ", err)
		return nil, qcloud.ErrInternalError
	}
	return model.CertInfoForShow(infos), ""
}

// HandleDomainCertDetail 获取域名证书
// @Summary 获取域名关联证书
// @Description 获取域名关联证书
// @Tags 监控管理
// @Accept mpfd
// @Produce json
// @Param action query string true "DescribeDomainCerts"
// @Param serviceType query string true "sslpod"
// @Param DomainId path string true "域名ID"
// @Success 200 {array} model.CertInfoShow "msg.Data"
// @Router /domains/cert/{DomainId} [get]
func HandleDomainCertDetail(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/domains/cert/:domainId").Inc()

	list, errCode := CoreCertDetail(req)
	if errCode != "" {
		return nil, errCode
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"Data":      list,
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}

// HandleDescribeDomainRegionalDetail 域名各地域详细检测信息
func HandleDescribeDomainRegionalDetail(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	domainRegionalDetail, errCode := coreDomainDetail(req)
	if errCode != "" {
		return nil, errCode
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"DomainRegionalDetail": domainRegionalDetail,
			"RequestId":            req.GetString("RequestId"),
		},
	}, ""
}

// DomainRegionalIPDetail 域名地域详细信息返回
type DomainRegionalIPDetail struct {
	IP            string
	Port          string
	Status        string
	CommonName    string
	Brand         string
	CertHash      string
	DNSNames      string
	Issuer        string
	CertBeginTime string
	CertEndTime   string
	IsAuto        bool
}

// DomainRegionalDetail 域名地域性检测结果
type DomainRegionalDetail struct {
	Region           string
	DetectionResults []DomainRegionalIPDetail
	DetectionTime    string
}

func coreDomainDetail(req qcloud.RequestBack) (interface{}, string) {
	a := req.GetValue("account").(*model.Account)

	domainID := req.GetInt("DomainId")
	if domainID < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	_, err := db.GetDomainAccountInfo(a.Id, domainID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, qcloud.ErrFailedOperation
		}
	}
	// 获取域名检测详情
	domainDetail, err := db.GetDomainAllRegionalResult(domainID)
	if err != nil {
		logrus.Error("coreDomainDetail.GetDomainAllRegionalResult: ", err)
		return nil, qcloud.ErrFailedOperation
	}
	// 获取uin域名的固定IP
	domainIPs, err := db.GetDomainIPsByUin(domainID, a.Uin)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			domainIPs = nil
		} else {
			domainIPs = nil
			logrus.Error("coreDomainDetail.GetDomainIPsByUin: ", err)
		}
	}
	var isAutoDetect = true
	var staticIPPorts = []model.IPPort{}
	var staticIPSet = map[string]struct{}{}
	if domainIPs != nil {
		isAutoDetect = domainIPs.IsAutoDetect
		err = json.Unmarshal([]byte(domainIPs.IpPorts), &staticIPPorts)
		if err != nil {
			logrus.Error("coreDomainDetail.Unmarshal: ", err)
		}
		for _, i := range staticIPPorts {
			staticIPSet[i.IP] = struct{}{}
		}
	}

	// 所有地域的检测结果集合
	// var drs = []model.DetectionResult{}
	// 所有地域的ip集合，去重
	// var IPSet = map[string]struct{}{}
	// 所有地域的检测结果集合（标准）
	var regionResults = []DomainRegionalDetail{}

	for _, r := range domainDetail {
		// 当前地域的检测结果
		var dr []model.DetectionResult
		// 当前地域的检测结果/标准
		var ipresults = []DomainRegionalIPDetail{}

		err := json.Unmarshal([]byte(r.DetectionResult), &dr)
		if err != nil {
			logrus.Error("coreDomainDetail.Unmarshal: ", err)
			continue
		}
		for _, i := range dr {
			// 如果没有开启自动检测，跳过非手动检测的IP
			if !isAutoDetect && !isContainIP(i.IP, staticIPSet) {
				continue
			}
			var certInfo = &model.CertInfoShow{}
			if len(i.Hashes) != 0 {
				certInfo, err = db.GetCertInfoByHash(i.Hashes[0])
			}

			if err != nil {
				logrus.Error("coreDomainDetail.GetCertInfoByHash: ", err)
				continue
			}
			ipresults = append(ipresults, DomainRegionalIPDetail{
				IP:            i.IP,
				Port:          i.Port,
				Status:        i.Status,
				CommonName:    certInfo.CN,
				Brand:         certInfo.Brand,
				CertHash:      certInfo.Hash,
				DNSNames:      certInfo.SANs,
				Issuer:        certInfo.Issuer,
				CertBeginTime: certInfo.BeginTime.Format("2006-01-02 15:04:05"),
				CertEndTime:   certInfo.EndTime.Format("2006-01-02 15:04:05"),
				IsAuto:        i.IsAuto,
			})
		}
		regionResults = append(regionResults, DomainRegionalDetail{
			Region:           r.Region,
			DetectionResults: ipresults,
			DetectionTime:    r.LastDetectionTime.Format("2006-01-02 15:04:05"),
		})
	}
	return regionResults, ""
}

// Paging 分页
type Paging struct {
	SearchTotal            int         // 搜索出来的数量
	Total                  int         // 总数
	Result                 interface{} // 结果
	AllowMonitoringCount   int         // 允许的监控数量
	CurrentMonitoringCount int         // 当前监控的数量
	AllowMaxAddDomain      int         // 允许添加域名总数
}

func isContainIP(ip string, autoIPs map[string]struct{}) bool {
	if _, ok := autoIPs[ip]; ok {
		return true
	}
	return false
}

// HandleDescribeDomainIPDetail 获得域名的IP监控详情
func HandleDescribeDomainIPDetail(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	domainIPDetail, errCode := coreDomainIPDetail(req)
	if errCode != "" {
		return nil, errCode
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"DomainId":     domainIPDetail.DomainId,
			"Domain":       domainIPDetail.Domain,
			"DefaultPort":  domainIPDetail.DefaultPort,
			"IPDetail":     domainIPDetail.IPDetail,
			"CertInfo":     domainIPDetail.CertInfo,
			"Tags":         domainIPDetail.Tags,
			"Grade":        domainIPDetail.Grade,
			"Status":       domainIPDetail.Status,
			"IsAutoDetect": domainIPDetail.IsAutoDetect,
			"RequestId":    req.GetString("RequestId"),
		},
	}, ""
}

// DomainIPDetailResponse 域名IP详细信息的返回格式
type DomainIPDetailResponse struct {
	DomainId     int
	Domain       string
	DefaultPort  string
	IPDetail     []*IPDetaliResponse
	CertInfo     model.CertInfoShow
	Tags         []string
	Grade        string
	Status       string
	IsAutoDetect bool
}

// IPDetaliResponse 每个IP的详细信息
type IPDetaliResponse struct {
	IP                string
	Port              string
	Status            string
	LastDetectionTime string
}

func coreDomainIPDetail(req qcloud.RequestBack) (*DomainIPDetailResponse, string) {
	a := req.GetValue("account").(*model.Account)

	domainID := req.GetInt("DomainId")
	if domainID < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	accountDomain, err := db.GetDomainAccountInfo(a.Id, domainID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, qcloud.ErrFailedOperation
		}
	}
	// 获取域名检测详情
	domainDetail, err := db.GetDomainAllRegionalResult(domainID)
	if err != nil {
		logrus.Error("coreDomainIPDetail.GetDomainAllRegionalResult: ", err)
		return nil, qcloud.ErrFailedOperation
	}
	// 获取uin域名的固定IP
	domainIPs, err := db.GetDomainIPsByUin(domainID, a.Uin)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			domainIPs = nil
		} else {
			domainIPs = nil
			logrus.Error("coreDomainIPDetail.GetDomainIPsByUin: ", err)
		}
	}
	var isAutoDetect = true
	var staticIPPorts = []model.IPPort{}
	var staticIPSet = map[string]struct{}{}
	if domainIPs != nil {
		isAutoDetect = domainIPs.IsAutoDetect
		err = json.Unmarshal([]byte(domainIPs.IpPorts), &staticIPPorts)
		if err != nil {
			logrus.Error("coreDomainIPDetail.Unmarshal: ", err)
		}
		for _, i := range staticIPPorts {
			staticIPSet[i.IP] = struct{}{}
		}
	}
	// 将每个地域的每个IP去重聚合
	var sdr = map[string]*IPDetaliResponse{}
	var hash string
	for _, r := range domainDetail {
		// 当前地域的检测结果
		var dr []model.DetectionResult

		err := json.Unmarshal([]byte(r.DetectionResult), &dr)
		if err != nil {
			logrus.Error("coreDomainIPDetail.Unmarshal: ", err)
			continue
		}
		for _, i := range dr {
			if hash == "" && len(i.Hashes) != 0 {
				hash = i.Hashes[0]
			}
			// 如果没有开启自动检测，跳过非手动检测的IP
			if !isAutoDetect && !isContainIP(i.IP, staticIPSet) {
				continue
			}
			// 非自动检测的IP，并且不在当前用户的固定IP列表里，跳过
			if !i.IsAuto && !isContainIP(i.IP, staticIPSet) {
				continue
			}
			// 如果已存在，跳过
			if _, ok := sdr[i.IP]; ok {
				continue
			}
			sdr[i.IP] = &IPDetaliResponse{
				IP:                i.IP,
				Port:              i.Port,
				Status:            i.Status,
				LastDetectionTime: r.LastDetectionTime.Format("2006-01-02 15:04:05"),
			}
		}
	}
	domainResult, err := db.GetDomainResultById(domainID)
	if err != nil {
		logrus.Error("coreDomainIPDetail.GetDomainResultById: ", err)
	}
	tags, err := db.GetTagsNameByDomainAccountId(accountDomain.Id)
	if err != nil {
		logrus.Error("coreDomainIPDetail.GetTagsNameByDomainAccountId: ", err)
	}
	certInfo, err := db.GetCertInfoByHash(hash)
	certInfo.TrustStatus = domainResult.TrustStatus
	if err != nil {
		logrus.Error("coreDomainIPDetail.GetCertInfoByHash: ", err)
	}
	hasNormal := false
	hasAbnormal := false
	ret := DomainIPDetailResponse{
		DomainId:    domainID,
		Domain:      domainResult.Domain,
		DefaultPort: domainResult.Port,
		IPDetail: func(sdr map[string]*IPDetaliResponse) []*IPDetaliResponse {
			var r = []*IPDetaliResponse{}
			for _, i := range sdr {
				r = append(r, i)
				if i.Status == common.CertTrust {
					hasNormal = true
				} else {
					hasAbnormal = true
				}
			}
			return r
		}(sdr),
		CertInfo: *certInfo,
		Tags:     tags,
		Grade:    domainResult.Grade,
	}
	if hasNormal == true && hasAbnormal == true {
		ret.Status = common.CertPartAbnormal
	} else if hasNormal == true && hasAbnormal == false {
		ret.Status = common.CertTrust
	}
	ret.IsAutoDetect = isAutoDetect
	return &ret, ""
}

// HandleDescribeDomainIPs 获取域名IP
func HandleDescribeDomainIPs(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	a := req.GetValue("account").(*model.Account)
	domainIDs := []int{}
	for _, i := range req.GetValue("DomainIds").([]interface{}) {
		domainIDs = append(domainIDs, int(i.(float64)))
	}
	ret := []interface{}{}
	for _, domainID := range domainIDs {
		// 如果没有关注这个域名，直接返回
		hasAttention, err := db.UserHaveAttentionDomain(a.Id, domainID)
		if err != nil || !hasAttention {
			return nil, qcloud.ErrFailedOperation
		}
		dr, err := db.GetDomainResultById(domainID)
		tempDomainIPInfos, IsAutoDetect := getDomainRegionalIPInfo(a, domainID, dr.Domain)
		temp := []model.DetectionResult{}
		for _, ipInfos := range tempDomainIPInfos {
			temp = append(temp, ipInfos)
		}
		ret = append(ret, map[string]interface{}{
			"DomainId":     domainID,
			"IPInfos":      temp,
			"IsAutoDetect": IsAutoDetect,
		})
	}
	return &qcloud.ResponseBack{
		Response: gin.H{
			"DomainIPs": ret,
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}

func getDomainRegionalIPInfo(a *model.Account, domainID int, domain string) (map[string]model.DetectionResult, bool) {
	domainRegionalInfo, err := db.GetDomainAllRegionalResult(domainID)
	if err != nil {
		logrus.Error("HandleDescribeDomainIPs.GetDomainAllRegionalResult ", err)
		return nil, true
	}
	staticIPSet, isAutoDetect := getStaticIPs(a, domainID)
	var ret = map[string]model.DetectionResult{}
	for _, i := range domainRegionalInfo {
		tempIPInfos := []model.DetectionResult{}
		json.Unmarshal([]byte(i.DetectionResult), &tempIPInfos)
		for _, j := range tempIPInfos {
			// 兼容域名为ip的情况
			if j.IP == domain {
				ret[j.IP] = j
				continue
			}
			// 如果域名是固定，且不在客户固定IP里，跳过
			if !isAutoDetect && !isContainIP(j.IP, staticIPSet) {
				continue
			}
			if !j.IsAuto && !isContainIP(j.IP, staticIPSet) {
				continue
			}
			if _, ok := ret[j.IP]; ok {
				continue
			}
			ret[j.IP] = j
		}
	}
	return ret, isAutoDetect
}

func getStaticIPs(a *model.Account, domainID int) (map[string]struct{}, bool) {
	// 获取uin域名的固定IP
	domainIPs, err := db.GetDomainIPsByUin(domainID, a.Uin)
	var staticIPPorts = []model.IPPort{}
	var staticIPSet = map[string]struct{}{}
	var isAutoDetect = true
	if err == nil {
		isAutoDetect = domainIPs.IsAutoDetect
		err = json.Unmarshal([]byte(domainIPs.IpPorts), &staticIPPorts)
		if err != nil {
			logrus.Error("coreDomainIPDetail.Unmarshal: ", err)
		}
		for _, i := range staticIPPorts {
			staticIPSet[i.IP] = struct{}{}
		}
	}
	return staticIPSet, isAutoDetect
}

// HandleDescribeDomainCertsDetail HandleDescribeDomainCertsDetail
func HandleDescribeDomainCertsDetail(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	// 获取入参
	a := req.GetValue("account").(*model.Account)
	domainId := req.GetInt("DomainId")
	// 判断该账号是否关注此域名
	exist, err := db.UserHaveAttentionDomain(a.Id, domainId)
	if err != nil {
		logrus.Error("HandleDescribeDomainCertsDetailreq.UserHaveAttentionDomain: ", err)
		return nil, qcloud.ErrFailedOperation
	}
	if !exist {
		return nil, qcloud.ErrInvalidDomain
	}
	// 获取域名基本监控信息
	domainResult, _ := db.GetDomainResultWithOtherInfo(a.Id, domainId)
	// 获取域名全地域证书信息
	domainAllIPCerts, ipstatus, isAutoDetect, _ := db.GetDomainAllIPCerts(a.Uin, domainId, domainResult.Domain)
	// 根据多IP的状态判断域名状态 优先展示高优先级
	minstatuscode := 10
	minstatusname := ""
	for j, i := range ipstatus {
		tempstatuscode := common.ChangeStatusToCode(i["Status"].(string))
		if tempstatuscode < minstatuscode {
			minstatuscode = tempstatuscode
			minstatusname = i["Status"].(string)
		}
		ipstatus[j]["Status"] = common.NewCertStatusText[ipstatus[j]["Status"].(string)]
	}
	domainResult.TrustStatus = minstatusname
	return &qcloud.ResponseBack{
		Response: gin.H{
			"DomainId":     domainId,
			"Domain":       domainResult.Domain,
			"DefaultPort":  443,
			"IPStatus":     ipstatus,
			"Status":       common.NewCertStatusText[domainResult.TrustStatus],
			"Tags":         domainResult.Tags,
			"Grade":        domainResult.Grade,
			"CertInfos":    domainAllIPCerts,
			"IsAutoDetect": isAutoDetect,
			"RequestId":    req.GetString("RequestId"),
		},
	}, ""
}
