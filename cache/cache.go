package cache

import (
	"github.com/miekg/dns"
)

type Cache interface {
	Add(*dns.Msg) error
	Get(*dns.Msg) (bool, *dns.Msg, error)
}
