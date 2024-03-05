package dns

const (
	headerSize = 12

	// Header.Bits
	BIT_QR = 1 << 15 // query/response (response=1)
	BIT_AA = 1 << 10 // authoritative
	BIT_TC = 1 << 9  // truncated
	BIT_RD = 1 << 8  // recursion desired
	BIT_RA = 1 << 7  // recursion available
)

const (
	// valid RR_Header.Rrtype and Question.qtype

	TypeNone  uint16 = 0
	TypeA     uint16 = 1
	TypeNS    uint16 = 2
	TypeCNAME uint16 = 5
	TypeSOA   uint16 = 6
	TypePTR   uint16 = 12
	TypeMX    uint16 = 15
	TypeTXT   uint16 = 16
	TypeAAAA  uint16 = 28
)

const (
	ClassINET = 1
)

const (
	RcodeSuccess        = 0
	RcodeFormatError    = 1
	RcodeServerFailure  = 2
	RcodeNameError      = 3
	RcodeNotImplemented = 4
	RcodeRefused        = 5
)

var TypeToRR = map[uint16]func() RR{
	TypeA:     func() RR { return new(A) },
	TypeAAAA:  func() RR { return new(AAAA) },
	TypeNS:    func() RR { return new(NS) },
	TypeCNAME: func() RR { return new(CNAME) },
	TypeSOA:   func() RR { return new(SOA) },
	TypePTR:   func() RR { return new(PTR) },
	TypeMX:    func() RR { return new(MX) },
	TypeTXT:   func() RR { return new(TXT) },
}
