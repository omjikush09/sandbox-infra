package ippool

import "sync"

var ipPool = IpPool{
	mu:    sync.Mutex{},
	Ideal: []*IP{},
	InUse: make(map[int]*IP),
}

func InitThePool() {
	ipPool.Init()
}

func GetAIP() (*IP, error) {
	return ipPool.LeaseIp()
}

func ReleaseIP(ip *IP) {
	ipPool.ReleaseIP(ip)
}
