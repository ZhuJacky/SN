package myconn

type ServerType int

const (
	Web ServerType = iota
	SMTP
	IMAP
	POP3
)

type CheckParams struct {
	Domain     string     //检测的域名
	Ip         string     //检测的ip
	Port       string     //检测的端口
	ServerType ServerType //检测的服务器类型
	MailDirect bool       //如果邮件服务器是否之间调用tls
}

//区分端口
func DetectionServerType(port string) ServerType {
	switch port {
	case "110":
		fallthrough
	case "995":
		return POP3

	case "25":
		fallthrough
	case "465":
		fallthrough
	case "587":
		fallthrough
	case "994":
		return SMTP

	case "143":
		fallthrough
	case "993":
		return IMAP

	default:
		return Web
	}

}

func ServerTypeToString(serverType ServerType) string {
	switch serverType {
	case Web:
		return "Web"
	case SMTP:
		return "SMTP"
	case IMAP:
		return "IMAP"
	case POP3:
		return "POP"
	default:
		return "Unknown"

	}
}
