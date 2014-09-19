package dnsforwarder

import (
	"encoding/json"
	"net"
	"testing"
)

func TestConfigurationJSONMarshalling(test *testing.T) {
	var err error

	startConfiguration := Configuration{}
	startConfiguration.ReadTimeout = 10
	startConfiguration.WriteTimeout = 10
	startConfiguration.NameServers = []net.TCPAddr{net.TCPAddr{IP: net.IPv4(208, 67, 222, 222), Port: 53}, net.TCPAddr{IP: net.IPv4(208, 67, 220, 220), Port: 53}}
	startConfiguration.TTL = 600

	test.Logf("Configuration Object:%v\n", startConfiguration)

	bytestartConfiguration, err := json.Marshal(startConfiguration)
	if err != nil {
		test.Error("Error Marshaling to JSON:" + err.Error())
	}

	test.Log("As JSON:" + string(bytestartConfiguration))

	endConfiguration := Configuration{}
	err = json.Unmarshal(bytestartConfiguration, &endConfiguration)
	if err != nil {
		test.Error("Error Unmarshaling to JSON:" + err.Error())
	}

	test.Logf("Configuration Object:%v\n", endConfiguration)

	if endConfiguration.ReadTimeout != startConfiguration.ReadTimeout {
		test.Error("Error: ReadTimeout Doesn't Match")
	}
	if endConfiguration.WriteTimeout != startConfiguration.WriteTimeout {
		test.Error("Error: WriteTimeout Doesn't Match")
	}
	if endConfiguration.TTL != startConfiguration.TTL {
		test.Error("Error: TTL Doesn't Match")
	}

	if len(endConfiguration.NameServers) != len(startConfiguration.NameServers) {
		test.Error("Error: NameServers Doesn't Match in Size?")
	}

	for i := 0; i < len(endConfiguration.NameServers); i++ {
		if endConfiguration.NameServers[i].String() != startConfiguration.NameServers[i].String() {
			test.Error("Error: NameServers Doesn't Match")
		}
	}

}

func TestConfigurationJSONUnmarshallingError(test *testing.T) {
	endConfiguration := Configuration{}

	//Broken JSON String
	err := json.Unmarshal([]byte(`{"ReadTimeout"}:10,"WriteTimeout":10,"NameServers":["208.67.222.222:53","208.67.220.220:53"],"TTL":600}`), &endConfiguration)
	if err == nil {
		test.Error("Expected JSON Error")
	} else {
		test.Logf("Error Message:%s\n", err.Error())
	}
}

func TestConfigurationJSONUnmarshallingNameserversError(test *testing.T) {
	endConfiguration := Configuration{}

	//Invalid DNS Server IP
	err := json.Unmarshal([]byte(`{"ReadTimeout":10,"WriteTimeout":10,"NameServers":["999.67.222.222:53","208.67.220.220:53"],"TTL":600}`), &endConfiguration)
	if err == nil {
		test.Error("Expected Unmashalling Error")
	} else {
		test.Logf("Error Message:%s\n", err.Error())
	}
}
