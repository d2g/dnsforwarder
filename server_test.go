package dnsforwarder

import (
	memorycache "github.com/d2g/dnsforwarder/cache/memory"
	memoryhosts "github.com/d2g/dnsforwarder/hosts/memory"
	"net"
	"testing"
)

/*
 * Example Server :D
 */
func Test_ExampleServer(test *testing.T) {
	configuration := Configuration{}
	configuration.ReadTimeout = 10
	configuration.WriteTimeout = 10
	configuration.TTL = 600
	//configuration.NameServers = []net.TCPAddr{net.TCPAddr{IP: net.IPv4(10, 187, 7, 101), Port: 53}, net.TCPAddr{IP: net.IPv4(10, 187, 7, 102), Port: 53}}
	configuration.NameServers = []net.TCPAddr{net.TCPAddr{IP: net.IPv4(192, 168, 1, 254), Port: 53}}

	cache := memorycache.Memory{}
	cache.Cache = make(map[string]memorycache.MemoryCacheRecord)

	hosts := memoryhosts.Memory{}
	hosts.Devices = make(map[string]net.IP)
	hosts.Add("raspberrypi", net.IPv4(192, 168, 1, 201))

	server := Server{}
	server.Configuration = &configuration
	server.Cache = &cache
	server.Hosts = &hosts

	go func() {
		err := server.ListenAndServeUDP(net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 53})
		if err != nil {
			test.Error("UDP Error:%v\n", err)
		}
	}()

	err := server.ListenAndServeTCP(net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 53})
	if err != nil {
		test.Error("TCP Error:%v\n", err)
	}
}
