package memory

import (
	"github.com/miekg/dns"
	"time"
)

type Memory struct {
	Cache map[string]MemoryCacheRecord
}

type MemoryCacheRecord struct {
	Expiry time.Time
	Record dns.Msg
}

func (this *Memory) Add(message *dns.Msg) error {
	if message.Question[0].Qtype == dns.TypeA && message.Question[0].Qclass == dns.ClassINET {
		this.Cache[message.Question[0].Name] = MemoryCacheRecord{Expiry: time.Now().Add(time.Duration(message.Answer[0].Header().Ttl) * time.Second), Record: *message}
	}
	return nil
}

func (this *Memory) Get(message *dns.Msg) (bool, *dns.Msg, error) {
	cachedMessage := this.Cache[message.Question[0].Name]
	if cachedMessage.Record.Id != 0 && cachedMessage.Expiry.After(time.Now()) {
		return true, &cachedMessage.Record, nil
	} else {
		empty := dns.Msg{}
		return false, &empty, nil
	}
}
