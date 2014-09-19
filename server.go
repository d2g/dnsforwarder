/*
Package dnsforwarder implements a DNS server with local lookup and caching.
Based on the DNS implementation by miekg (github.com/miekg/dns).
*/

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

/*
The Implementation of the Hijacker interface must write the message to the writer.
It returns weather the message was hijacked (true) or not (false). With any errors.
*/
type Hijacker func(dns.ResponseWriter, *dns.Msg) (bool, error)

// Listen for incoming TCP Connections
func (this *Server) ListenAndServeTCP(address net.TCPAddr) error {

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", this.tCPRequest)

	tcpServer := &dns.Server{Addr: address.String(),
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  (time.Duration(this.Configuration.ReadTimeout) * time.Second),
		WriteTimeout: (time.Duration(this.Configuration.WriteTimeout) * time.Second)}

	err := tcpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

// Listen for incoming UDP Connections
func (this *Server) ListenAndServeUDP(address net.UDPAddr) error {

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", this.uDPRequest)

	udpServer := &dns.Server{Addr: address.String(),
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  (time.Duration(this.Configuration.ReadTimeout) * time.Second),
		WriteTimeout: (time.Duration(this.Configuration.WriteTimeout) * time.Second)}

	err := udpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

//Wrapper to turn TCP requests into generic requests.
func (this *Server) tCPRequest(response dns.ResponseWriter, message *dns.Msg) {
	this.request("tcp", response, message)
}

//Wrapper to turn UDP requests into generic requests.
func (this *Server) uDPRequest(response dns.ResponseWriter, message *dns.Msg) {
	this.request("udp", response, message)
}

//Handles the request writing the response back:
//Checked in the following order:
//	* Hijacker
//  * Local Hosts
//  * Cache
//  * Remote Lookup
func (this *Server) request(method string, response dns.ResponseWriter, message *dns.Msg) {
	//Hijack Request If Needed.
	if this.Hijacker != nil {
		wasHijacked, err := this.Hijacker(response, message)
		if err != nil {
			log.Printf("Error: DNS Hijack:%v\n", err)
		}
		if wasHijacked {
			return
		}
	}

	//Check Against Local Host File
	if this.Hosts != nil {
		isLocal, ip, err := this.Hosts.Get(strings.TrimSuffix(message.Question[0].Name, "."))

		if err != nil {
			log.Println("Error: Looking At Local DNS:" + err.Error())

		} else {
			if isLocal {
				//Return the local device detail
				localResponse := new(dns.Msg)
				localResponse.SetReply(message)
				rr_header := dns.RR_Header{Name: message.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: this.Configuration.TTL}

				a := &dns.A{Hdr: rr_header, A: ip}
				localResponse.Answer = append(localResponse.Answer, a)
				err := response.WriteMsg(localResponse)
				if err != nil {
					log.Printf("Error: Writing Local Response:%v\n", err)
				}

				return
			}
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
					log.Printf("Error: Writing Cached Response:%v\n", err)
				}
				return
			}
		} else {
			log.Println("Warning: Cache Error \"" + err.Error() + "\" On " + strings.TrimSuffix(message.Question[0].Name, "."))
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
		log.Printf("Error: Writing Response:%v\n", err)
	}

	//Cache Response :D
	if this.Cache != nil {
		err = this.Cache.Add(result)
		if err != nil {
			log.Printf("Error: Adding Response To Cache:%v\n", err)
		}
	}
}

//Does the remote lookup in the remote DNS servers.
func (this *Server) remoteDNSLookup(protocol string, request *dns.Msg) (*dns.Msg, error) {
	dnsClient := &dns.Client{
		Net:          protocol,
		ReadTimeout:  (time.Duration(this.Configuration.ReadTimeout) * time.Second),
		WriteTimeout: (time.Duration(this.Configuration.ReadTimeout) * time.Second),
	}

	var err error = nil

	for _, nameserver := range this.Configuration.NameServers {
		result, _, err := dnsClient.Exchange(request, nameserver.String())

		if err != nil {
			log.Printf("Error: DNS:" + err.Error())
			continue
		}

		if result != nil && result.Rcode != dns.RcodeSuccess {
			continue
		}

		return result, nil
	}

	if err != nil {
		return nil, err
	} else {
		return nil, errors.New("Error: Unknown DNS Error")
	}
}
