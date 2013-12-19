package memory

import (
	"net"
)

type Memory struct {
	Devices map[string]net.IP
}

func (this *Memory) Add(hostname string, ip net.IP) error {
	this.Devices[hostname] = ip
	return nil
}

func (this *Memory) Get(hostname string) (bool, net.IP, error) {
	ip := this.Devices[hostname]

	if ip != nil {
		return true, ip, nil
	} else {
		return false, ip, nil
	}
}
