// Package qcloud provides ...
package qcloud

// qcloud error
const (
	ErrInvalidParameter      = "InvalidParameter"
	ErrInvalidParameterValue = "InvalidParameterValue"
	ErrMissingParameter      = "MissingParameter"
	ErrUnknownParameter      = "UnknownParameter"
	ErrAuthFailure           = "AuthFailure"
	ErrInternalError         = "InternalError"
	ErrInvalidAction         = "InvalidAction"
	ErrUnauthorizedOperation = "UnauthorizedOperation"
	ErrRequestLimitExceeded  = "RequestLimitExceeded"
	ErrNoSuchVersion         = "NoSuchVersion"
	ErrUnsupportedRegion     = "UnsupportedRegion"
	ErrUnsupportedOperation  = "UnsupportedOperation"
	ErrResourceNotFound      = "ResourceNotFound"
	ErrLimitExceepted        = "LimitExceeded"
	ErrResourceUnavailable   = "ResourceUnavailable"
	ErrResourceInsufficient  = "ResourceInsufficient"
	ErrFailedOperation       = "FailedOperation"
	ErrResourceInUse         = "ResourceInUse"
	ErrDryRunOperation       = "DryRunOperation"
	// Custom error
	ErrInvalidServerType    = "InvalidParameter.InvalidServerType"
	ErrInvalidPort          = "InvalidParameter.InvalidPort"
	ErrInvalidDomain        = "InvalidParameter.InvalidDomain"
	ErrInvalidIP            = "InvalidParameter.InvalidIP"
	ErrInvalidTagName       = "InvalidParameter.InvalidTagName"
	ErrTooManyTag           = "InvalidParameter.TooManyTag"
	ErrLimitedAddDomain     = "LimitExceeded.AddExceeded"
	ErrLimitedMonitorDomain = "LimitExceeded.MonitorExceeded"
	ErrRepetitionAdd        = "FailedOperation.RepetitionAdd"
	ErrInvalidSearchType    = "InvalidParameterValue.InvalidSearchType"
	ErrFailedResolveDomain  = "FailedOperation.ResolveDomainFailed"
	ErrOtherPlanInUse       = "ErrFailedOperation.OtherPlanInUse"

	ErrInvalidNoticeType = "InvalidParameterValue.InvalidNoticeType"

	ErrProductNotFound = "ResourceNotFound.Product"
)

var ErrDesc = map[string]string{
	ErrInvalidParameter:      "参数错误（包括参数格式、类型等错误）",
	ErrInvalidParameterValue: "参数取值错误",
	ErrMissingParameter:      "缺少参数错误，必传参数没填",
	ErrUnknownParameter:      "未知参数错误，用户多传未定义的参数会导致错误",
	ErrAuthFailure:           "CAM签名/鉴权错误",
	ErrInternalError:         "内部错误",
	ErrInvalidAction:         "接口不存在",
	ErrUnauthorizedOperation: "未授权操作",
	ErrRequestLimitExceeded:  "请求的次数超过了频率限制",
	ErrNoSuchVersion:         "接口版本不存在",
	ErrUnsupportedRegion:     "接口不支持所传地域",
	ErrUnsupportedOperation:  "操作不支持",
	ErrResourceNotFound:      "资源不存在",
	ErrLimitExceepted:        "超过配额限制",
	ErrResourceUnavailable:   "资源不可用",
	ErrResourceInsufficient:  "资源不足",
	ErrFailedOperation:       "操作失败",
	ErrResourceInUse:         "资源被占用",
	ErrDryRunOperation:       "DryRun操作，代表请求将会是成功的，只是多传了DryRun参数",
	// Custom error
	ErrInvalidServerType:    "无效的监控类型",
	ErrInvalidPort:          "无效的端口",
	ErrInvalidDomain:        "无效的域名",
	ErrInvalidIP:            "无效的IP",
	ErrInvalidTagName:       "tag只能包含中文、英文、数字且在10个字符以内",
	ErrTooManyTag:           "tag最多添加3个",
	ErrLimitedAddDomain:     "监控超出限制",
	ErrLimitedMonitorDomain: "监控告警超出限制",
	ErrRepetitionAdd:        "重复添加",
	ErrInvalidSearchType:    "无效的搜索类型",
	ErrFailedResolveDomain:  "解析域名失败",
	ErrOtherPlanInUse:       "您已购买过其它套餐",

	ErrInvalidNoticeType: "无效的通知类型",

	ErrProductNotFound: "产品不存在",
}
