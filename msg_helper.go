package dns

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func unpackDataA(msg []byte, off int) (net.IP, int, error) {
	if off+net.IPv4len > len(msg) {
		return nil, len(msg), fmt.Errorf("overflow unpacking a")
	}
	return CloneSlice(msg[off : off+net.IPv4len]), off + net.IPv4len, nil
}

func unpackDataAAAA(msg []byte, off int) (net.IP, int, error) {
	if off+net.IPv6len > len(msg) {
		return nil, len(msg), fmt.Errorf("overflow unpacking a")
	}
	return CloneSlice(msg[off : off+net.IPv6len]), off + net.IPv6len, nil
}

func packDataA(a net.IP, msg []byte, off int) (off1 int, err error) {
	switch len(a) {
	case net.IPv4len, net.IPv6len:
		if off+net.IPv4len > len(msg) {
			return len(msg), fmt.Errorf("overflow packing a")
		}
		copy(msg[off:], a.To4())
		off += net.IPv4len
	case 0:
		// allowed
	default:
		return len(msg), fmt.Errorf("overflow packing a")
	}

	return off, nil
}

func packDataAAAA(a net.IP, msg []byte, off int) (off1 int, err error) {
	switch len(a) {
	case net.IPv6len:
		if off+net.IPv4len > len(msg) {
			return len(msg), fmt.Errorf("overflow packing a")
		}
		copy(msg[off:], a.To4())
		off += net.IPv6len
	case 0:
		// allowed
	default:
		return len(msg), fmt.Errorf("overflow packing a")
	}

	return off, nil
}

func packDomainName(name string, msg []byte, off int, compression map[string]uint16) (int, error) {
	var begin int

	pointer := -1
loop:
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch c {
		case '.':
			labelLen := i - begin

			// find/store pointer
			if !isRootLabel(name, begin, len(name)) {
				prefix := name[begin:]
				if p, ok := compression[prefix]; ok {
					pointer = int(p)
					break loop
				} else {
					compression[name[begin:]] = uint16(off)
				}
			}

			// store label
			msg[off] = byte(labelLen)
			ss := name[begin:i]
			_ = ss
			copy(msg[off+1:], name[begin:i])

			off += 1 + labelLen
			begin = i + 1
		}
	}

	if isRootLabel(name, 0, len(name)) {
		return off, nil
	}

	if pointer != -1 {
		binary.BigEndian.PutUint16(msg[off:], uint16(pointer|0xC000))
		return off + 2, nil
	}

	msg[off] = 0
	return off + 1, nil
}

func isRootLabel(s string, off, end int) bool {
	return s[off:end] == "."
}

func unpackDomainName(buf []byte, off int) (string, int, error) {
	// TODO
	var s []byte
	ptr := 0
	off1 := 0
loop:
	for {
		c := int(buf[off])
		off++
		switch c & 0xC0 {
		case 0x00:
			// normal labels
			if c == 0x00 {
				break loop
			}
			for _, b := range buf[off : off+c] {
				s = append(s, b)
			}
			s = append(s, '.')
			off += c
		case 0xC0:
			// pointer
			c1 := int(buf[off])
			off++
			if ptr == 0 {
				off1 = off
			}
			if ptr++; ptr > 126 {
				return "", 0, fmt.Errorf("infinite loop")
			}
			off = c1 & 0x3F
		}
	}
	if ptr == 0 {
		off1 = off
	}

	if len(s) == 0 {
		return ".", off1, nil
	}

	return string(s), off1, nil
}

func packUint16(i uint16, buf []byte, off int) (int, error) {
	if off+2 > len(buf) {
		return len(buf), fmt.Errorf("overflow packing uint16")
	}

	binary.BigEndian.PutUint16(buf[off:], i)
	return off + 2, nil
}

func packUint32(i uint32, buf []byte, off int) (int, error) {
	if off+4 > len(buf) {
		return len(buf), fmt.Errorf("overflow packing uint32")
	}

	binary.BigEndian.PutUint32(buf[off:], i)
	return off + 4, nil
}

func packRRSlice(rrs []RR, buf []byte, off int, compression map[string]uint16) (off1 int, err error) {
	for _, rr := range rrs {
		off, err = packDomainName(rr.Header().Name, buf, off, compression)
		if err != nil {
			return off, err
		}

		off, err = packUint16(rr.Header().Rrtype, buf, off)
		if err != nil {
			return off, err
		}

		off, err = packUint16(rr.Header().Class, buf, off)
		if err != nil {
			return off, err
		}

		off, err = packUint32(rr.Header().Ttl, buf, off)
		if err != nil {
			return off, err
		}

		// set rdLength later
		offAtRdLengh := off
		off += 2

		off, err = rr.pack(buf, off, compression)
		if err != nil {
			return off, err
		}

		rdLength := off - offAtRdLengh - 2
		_, err = packUint16(uint16(rdLength), buf, offAtRdLengh)
		if err != nil {
			return off, err
		}
	}

	return off, nil
}

func unpackUint16(buf []byte, off int) (uint16, int, error) {
	if off+2 > len(buf) {
		return 0, len(buf), fmt.Errorf("overflow unpacking uint16")
	}
	return binary.BigEndian.Uint16(buf[off:]), off + 2, nil
}

func unpackUint32(buf []byte, off int) (uint32, int, error) {
	if off+4 > len(buf) {
		return 0, len(buf), fmt.Errorf("overflow unpacking uint32")
	}
	return binary.BigEndian.Uint32(buf[off:]), off + 4, nil
}

func unpackRRSlice(data []byte, off int, count int) ([]RR, int, error) {
	var err error
	var res []RR
	// rr
	for i := 0; i < count; i++ {
		var rh RR_Header
		rh.Name, off, err = unpackDomainName(data, off)
		if err != nil {
			return nil, off, err
		}
		rh.Rrtype, off, err = unpackUint16(data, off)
		if err != nil {
			return nil, off, err
		}
		rh.Class, off, err = unpackUint16(data, off)
		if err != nil {
			return nil, off, err
		}
		rh.Ttl, off, err = unpackUint32(data, off)
		if err != nil {
			return nil, off, err
		}
		rh.Rdlength, off, err = unpackUint16(data, off)
		if err != nil {
			return nil, off, err
		}

		end := off + int(rh.Rdlength)

		var rr RR
		rr, off, err = unpackRR(rh, data, off)
		if err != nil {
			return nil, off, err
		}

		if end != off {
			return nil, 0, fmt.Errorf("bad rdlength")
		}

		if rr != nil {
			res = append(res, rr)
		}
	}

	return res, off, nil
}

func unpackRR(rh RR_Header, data []byte, off int) (RR, int, error) {
	var err error

	var rr RR
	if rrFunc, ok := TypeToRR[rh.Rrtype]; ok {
		rr = rrFunc()
		*rr.Header() = rh
	} else {
		log.Printf("unsupported rr type %d", rh.Rrtype)
		return nil, off + int(rh.Rdlength), nil
	}

	if rh.Rdlength == 0 {
		return rr, off, nil
	}

	off, err = rr.unpack(data, off)
	if err != nil {
		return nil, off, err
	}

	return rr, off, nil
}

func unpackTxt(msg []byte, off int, end int) ([]string, int, error) {
	var txt []string

	for off < end {
		l := int(msg[off])
		off++
		t := string(msg[off : off+l])
		txt = append(txt, t)
		off += l
	}

	if off != end {
		return nil, 0, fmt.Errorf("offset not equal to end")
	}

	return txt, off, nil
}

func packTxt(txt []string, msg []byte, off int) (off1 int, err error) {
	for _, t := range txt {
		l := len(t)
		msg[off] = uint8(l)
		off++
		copy(msg[off:], t)
		off += l
	}

	return off, nil
}
