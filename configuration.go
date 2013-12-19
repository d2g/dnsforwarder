package dnsforwarder

import (
	"net"
)

//const OPENDNS_PRIMARY string = "208.67.222.222"
//const OPENDNS_SECONDARY string = "208.67.220.220"

type Configuration struct {
	ReadTimeout  int
	WriteTimeout int
	NameServers  []net.TCPAddr
	TTL          uint32
}
