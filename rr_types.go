package dns

import (
	"net"
)

type A struct {
	Hdr RR_Header
	A   net.IP
}

type AAAA struct {
	Hdr  RR_Header
	AAAA net.IP
}

type CNAME struct {
	Hdr    RR_Header
	Target string
}

type NS struct {
	Hdr RR_Header
	Ns  string
}

type TXT struct {
	Hdr RR_Header
	Txt []string
}

type MX struct {
	Hdr        RR_Header
	Preference uint16
	Exchange   string
}

type SOA struct {
	Hdr     RR_Header
	Mname   string
	Rname   string
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	MinTtl  uint32
}

type PTR struct {
	Hdr           RR_Header
	PtrDomainName string
}
