package hosts

import (
	"net"
)

type Hosts interface {
	Add(string, net.IP) error
	Get(string) (bool, net.IP, error)
}
