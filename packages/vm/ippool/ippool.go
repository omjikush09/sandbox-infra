package ippool

import (
	"errors"
	"fmt"
	"sync"
)

type IPStatus string

const (
	IPFree     IPStatus = "ideal"
	IPLeased   IPStatus = "leased"
	IPDraining IPStatus = "draining"
)

const MaxSlots = 243

type IP struct {
	ID     int
	state  IPStatus
	HostIP string
	VmIP   string
}

type IpPool struct {
	mu    sync.Mutex
	Ideal []*IP
	InUse map[int]*IP
}

func (p *IpPool) Init() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.InUse == nil {
		p.InUse = make(map[int]*IP)
	}
	p.Ideal = p.Ideal[:0]
	clear(p.InUse)

	for i := 1; i <= MaxSlots; i++ {
		vm := IP{
			ID:     i,
			state:  IPFree,
			HostIP: fmt.Sprintf("172.16.%d.1", i),
			VmIP:   fmt.Sprintf("172.16.%d.2", i),
		}
		p.Ideal = append(p.Ideal, &vm)
	}
}

func (p *IpPool) LeaseIp() (*IP, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Ideal) == 0 {
		return nil, errors.New("No IP available")
	}
	leaseIP := p.Ideal[0]
	p.Ideal = p.Ideal[1:]
	leaseIP.state = IPLeased
	p.InUse[leaseIP.ID] = leaseIP
	return leaseIP, nil
}

func (p *IpPool) ReleaseIP(ip *IP) error {
	if ip == nil {
		return fmt.Errorf("IP not found")
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.InUse[ip.ID]; !ok {
		return nil
	}

	delete(p.InUse, ip.ID)
	ip.state = IPFree
	p.Ideal = append(p.Ideal, ip)
	return nil
}
