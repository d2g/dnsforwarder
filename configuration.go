package dnsforwarder

import (
	"encoding/json"
	"net"
)

type Configuration struct {
	ReadTimeout  int
	WriteTimeout int
	NameServers  []net.TCPAddr
	TTL          uint32
}

func (t Configuration) MarshalJSON() ([]byte, error) {
	nameServers := make([]string, 0)

	for _, value := range t.NameServers {
		nameServers = append(nameServers, value.String())
	}

	stringMarshal := struct {
		ReadTimeout  int
		WriteTimeout int
		NameServers  []string
		TTL          uint32
	}{
		t.ReadTimeout,
		t.WriteTimeout,
		nameServers,
		t.TTL,
	}

	return json.Marshal(stringMarshal)
}

func (t *Configuration) UnmarshalJSON(data []byte) (err error) {
	stringUnMarshal := struct {
		ReadTimeout  int
		WriteTimeout int
		NameServers  []string
		TTL          uint32
	}{}

	err = json.Unmarshal(data, &stringUnMarshal)
	if err != nil {
		return
	}

	t.ReadTimeout = stringUnMarshal.ReadTimeout
	t.WriteTimeout = stringUnMarshal.WriteTimeout
	t.TTL = stringUnMarshal.TTL

	nameServers := make([]net.TCPAddr, 0)

	for _, value := range stringUnMarshal.NameServers {
		var address *net.TCPAddr
		address, err = net.ResolveTCPAddr("tcp4", value)
		if err != nil {
			break
		} else {
			nameServers = append(nameServers, *address)
		}
	}

	t.NameServers = nameServers

	return
}
