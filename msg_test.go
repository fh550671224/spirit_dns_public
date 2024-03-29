package dns

import (
	"context"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"testing"
)

func TestPack(t *testing.T) {
	msg := Msg{
		MsgHdr: MsgHdr{
			Id:                 51158,
			Response:           true,
			Opcode:             0,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   true,
			RecursionAvailable: true,
			Zero:               false,
			Rcode:              0,
		},
		Question: []Question{
			{
				Name:   "www.baidu.com.",
				QType:  1,
				QClass: 1,
			},
		},
		Answer: []RR{
			&CNAME{
				Hdr: RR_Header{
					Name:     "www.baidu.com.",
					Rrtype:   5,
					Class:    1,
					Ttl:      1200,
					Rdlength: 18,
				},
				Target: "www.a.shifen.com.",
			},
			&A{
				Hdr: RR_Header{
					Name:     "www.a.shifen.com.",
					Rrtype:   1,
					Class:    1,
					Ttl:      120,
					Rdlength: 4,
				},
				A: net.IP{
					36, 155, 132, 3,
				},
			},
		},
	}

	_msg := dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:                 51158,
			Response:           true,
			Opcode:             0,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   true,
			RecursionAvailable: true,
			Zero:               false,
			Rcode:              0,
		},
		Question: []dns.Question{
			{
				Name:   "www.baidu.com.",
				Qtype:  1,
				Qclass: 1,
			},
		},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:     "www.baidu.com.",
					Rrtype:   5,
					Class:    1,
					Ttl:      1200,
					Rdlength: 18,
				},
				Target: "www.a.shifen.com.",
			},
			&dns.A{
				Hdr: dns.RR_Header{
					Name:     "www.a.shifen.com.",
					Rrtype:   1,
					Class:    1,
					Ttl:      120,
					Rdlength: 4,
				},
				A: net.IP{
					36, 155, 132, 3,
				},
			},
		},
	}
	_ = msg
	_ = _msg

	data, err := msg.Pack()
	if err != nil {
		t.Error(err)
	}
	t.Log(data)

	checkMsg := new(dns.Msg)
	err = checkMsg.Unpack(data)
	if err != nil {
		log.Printf("dns.UnPack err: %v\n", err)
	}

	_data, err := _msg.Pack()
	if err != nil {
		t.Error(err)
	}
	t.Log(_data)

}

func TestTtt(t *testing.T) {
	a := "1,2,2,3,4,5,"
	aa := strings.Split(a, ",")

	t.Log(aa)
}

func TestGetRedisCacheByKey(t *testing.T) {
	var c RedisClient
	ctx := context.Background()
	c.InitRedis(ctx, "localhost:6379")

	c.GetRedisCacheByKey(ctx, Question{
		Name:   "baidu.com",
		QType:  1,
		QClass: 1,
	})
}

func TestPackTxt(t *testing.T) {
	txt := []string{
		"111",
		"2222",
		"3333",
	}

	buf := make([]byte, 100)

	i, err := packTxt(txt, buf, 0)
	if err != nil {
		return
	}
	t.Log(i)
}
func TestUnpackTxt(t *testing.T) {

	buf := []byte{3, 49, 49, 49, 4, 49, 49, 49, 49}

	txt, _, err := unpackTxt(buf, 0, 9)
	if err != nil {
		return
	}
	t.Log(txt)
}

func TestPackDataAAAA(t *testing.T) {
	a := net.ParseIP("2607:f8b0:400a:804::200e")
	b := make([]byte, 100)
	packDataAAAA(a, b, 0)
}
