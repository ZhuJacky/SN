package etcd

// 预设的排除的
var (
	DefaultExcludeInstanceTypes = []string{}
	DefaultExcludeInstanceIPs   = []string{}
)

// GetAvailableIntranetAddr 获取可用的内网IP
func GetAvailableIntranetAddr(appName AppName) []string {
	lock.RLock()
	defer lock.RUnlock()

	infos := AppInfos[appName]
	availableAddr := []string{}

Loop:
	for _, info := range infos {

		for _, ip := range DefaultExcludeInstanceIPs {
			if ip == info.IntranetAddr {
				continue Loop
			}
		}

		for _, insType := range DefaultExcludeInstanceTypes {
			if info.InstanceType == insType {
				continue Loop
			}
		}

		availableAddr = append(availableAddr, info.IntranetAddr)
	}
	return availableAddr
}
