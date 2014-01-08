package dnsforwarder

import (
	"errors"
	"github.com/d2g/dnsforwarder/cache"
	"github.com/d2g/dnsforwarder/hosts"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	Configuration *Configuration
	Cache         cache.Cache
	Hosts         hosts.Hosts
	Hijacker
}

// The Hijacker must write the message to the writer.
/*
 * Return: Bool (Was it hijacked?) if True then the client should have written to the DNS Response Writer
 * Return: Any Errors?
 */
type Hijacker func(dns.ResponseWriter, *dns.Msg) (bool, error)

func (this *Server) ListenAndServeTCP(address net.TCPAddr) error {

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", this.TCPRequest)

	tcpServer := &dns.Server{Addr: address.String(),
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  (time.Duration(this.Configuration.ReadTimeout) * time.Second),
		WriteTimeout: (time.Duration(this.Configuration.WriteTimeout) * time.Second)}

	err := tcpServer.ListenAndServe()
	if err != nil {
		log.Fatal("TCP DNS Server Failed, Error:" + err.Error())
		return err
	}

	return nil
}

func (this *Server) ListenAndServeUDP(address net.UDPAddr) error {

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", this.UDPRequest)

	udpServer := &dns.Server{Addr: address.String(),
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  (time.Duration(this.Configuration.ReadTimeout) * time.Second),
		WriteTimeout: (time.Duration(this.Configuration.WriteTimeout) * time.Second)}

	err := udpServer.ListenAndServe()
	if err != nil {
		log.Println("UDP DNS Server Failed, Error:" + err.Error())
		return err
	}

	return nil
}

func (this *Server) TCPRequest(response dns.ResponseWriter, message *dns.Msg) {
	this.Request("tcp", response, message)
}

func (this *Server) UDPRequest(response dns.ResponseWriter, message *dns.Msg) {
	this.Request("udp", response, message)
}

func (this *Server) Request(method string, response dns.ResponseWriter, message *dns.Msg) {
	//Hijack Request If Needed.
	if this.Hijacker != nil {
		wasHijacked, err := this.Hijacker(response, message)
		if err != nil {
			log.Printf("DNS Hijack Error:%v\n", err)
		}
		if wasHijacked {
			return
		}
	}

	//Lets check against the cache
	if this.Cache != nil {
		hitCache, cacheMessage, err := this.Cache.Get(message)
		if err == nil {
			//No Error Getting Cache
			if hitCache {
				cacheMessage.Id = message.Id
				err := response.WriteMsg(cacheMessage)
				if err != nil {
					log.Printf("Error Writing Cached Response:%v\n", err)
				}
				return
			}
		} else {
			log.Println("Warning Cache Error:" + err.Error())
		}
	}

	//Check Against Local Host File
	if this.Hosts != nil {
		isLocal, ip, err := this.Hosts.Get(strings.TrimSuffix(message.Question[0].Name, "."))

		if err != nil {
			log.Println("Error Looking At Local DNS:" + err.Error())

		} else {
			if isLocal {
				//Return the local device detail
				localResponse := new(dns.Msg)
				localResponse.SetReply(message)
				rr_header := dns.RR_Header{Name: message.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: this.Configuration.TTL}

				a := &dns.A{rr_header, ip}
				localResponse.Answer = append(localResponse.Answer, a)
				err := response.WriteMsg(localResponse)
				if err != nil {
					log.Printf("Error Writing Local Response:%v\n", err)
				}

				return
			}
		}
	}

	//So we didn't find the DNS locally.
	result, err := this.remoteDNSLookup(method, message)
	if err != nil {
		dns.HandleFailed(response, message)
		return
	}

	err = response.WriteMsg(result)
	if err != nil {
		log.Printf("Error Writing Response:%v\n", err)
	}

	//Cache Response :D
	if this.Cache != nil {
		err = this.Cache.Add(result)
		if err != nil {
			log.Printf("Error Adding Response To Cache:%v\n", err)
		}
	}
}

func (this *Server) remoteDNSLookup(protocol string, request *dns.Msg) (*dns.Msg, error) {
	dnsClient := &dns.Client{
		Net:          protocol,
		ReadTimeout:  (time.Duration(this.Configuration.ReadTimeout) * time.Second),
		WriteTimeout: (time.Duration(this.Configuration.ReadTimeout) * time.Second),
	}

	var err error = nil

	for _, nameserver := range this.Configuration.NameServers {
		//result, returnTime, err := dnsClient.Exchange(request, nameserver)
		result, _, err := dnsClient.Exchange(request, nameserver.String())

		if err != nil {
			log.Printf("DNS Error:" + err.Error())
			continue
		}

		if result != nil && result.Rcode != dns.RcodeSuccess {
			//log.Printf("%s failed to get an valid answer on %s\n", request.Question[0].Name, nameserver)
			continue
		}

		//if dns.IsFqdn(request.Question[0].Name) {
		//	log.Printf("%s resolv on %s ttl: %d\n", request.Question[0].Name[:len(request.Question[0].Name)-1], nameserver, returnTime)
		//} else {
		//	log.Printf("%s resolv on %s ttl: %d\n", request.Question[0].Name, nameserver, returnTime)
		//}

		return result, nil
	}

	if err != nil {
		return nil, err
	} else {
		return nil, errors.New("Unknown DNS Error")
	}
}
