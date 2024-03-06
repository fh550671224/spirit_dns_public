package dns

func (rr *A) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *AAAA) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *CNAME) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *NS) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *MX) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *SOA) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *PTR) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *TXT) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *A) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}

func (rr *AAAA) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}

func (rr *CNAME) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}

func (rr *NS) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}
func (rr *MX) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}
func (rr *SOA) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}
func (rr *PTR) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}

func (rr *TXT) len() (len int) {
	return rr.Header().len() + int(rr.Header().Rdlength)
}

func (rr *A) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packDataA(rr.A, msg, off)
	if err != nil {
		return off, err
	}
	return off, nil
}

func (rr *A) unpack(msg []byte, off int) (off1 int, err error) {
	rr.A, off, err = unpackDataA(msg, off)
	if err != nil {
		return off, err
	}
	return off, nil
}

func (rr *AAAA) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packDataAAAA(rr.AAAA, msg, off)
	if err != nil {
		return off, err
	}
	return off, nil
}

func (rr *AAAA) unpack(msg []byte, off int) (off1 int, err error) {
	rr.AAAA, off, err = unpackDataAAAA(msg, off)
	if err != nil {
		return off, err
	}
	return off, nil
}

func (rr *CNAME) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packDomainName(rr.Target, msg, off, compression)
	if err != nil {
		return off, err
	}

	return off, nil
}

func (rr *CNAME) unpack(msg []byte, off int) (off1 int, err error) {
	rr.Target, off, err = unpackDomainName(msg, off)
	if err != nil {
		return off, err
	}
	return off, nil
}

func (rr *NS) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packDomainName(rr.Ns, msg, off, compression)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *NS) unpack(msg []byte, off int) (off1 int, err error) {
	rr.Ns, off, err = unpackDomainName(msg, off)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *TXT) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packTxt(rr.Txt, msg, off)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *TXT) unpack(msg []byte, off int) (off1 int, err error) {
	end := off + int(rr.Header().Rdlength)
	rr.Txt, off, err = unpackTxt(msg, off, end)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *PTR) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packDomainName(rr.PtrDomainName, msg, off, compression)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *PTR) unpack(msg []byte, off int) (off1 int, err error) {
	rr.PtrDomainName, off, err = unpackDomainName(msg, off)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *SOA) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packDomainName(rr.Mname, msg, off, compression)
	if err != nil {
		return 0, err
	}

	off, err = packDomainName(rr.Rname, msg, off, compression)
	if err != nil {
		return 0, err
	}

	off, err = packUint32(rr.Serial, msg, off)
	if err != nil {
		return 0, err
	}
	off, err = packUint32(rr.Refresh, msg, off)
	if err != nil {
		return 0, err
	}
	off, err = packUint32(rr.Retry, msg, off)
	if err != nil {
		return 0, err
	}
	off, err = packUint32(rr.Expire, msg, off)
	if err != nil {
		return 0, err
	}
	off, err = packUint32(rr.MinTtl, msg, off)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *SOA) unpack(msg []byte, off int) (off1 int, err error) {
	rr.Mname, off, err = unpackDomainName(msg, off)
	if err != nil {
		return 0, err
	}
	rr.Rname, off, err = unpackDomainName(msg, off)
	if err != nil {
		return 0, err
	}

	rr.Serial, off, err = unpackUint32(msg, off)
	if err != nil {
		return 0, err
	}
	rr.Refresh, off, err = unpackUint32(msg, off)
	if err != nil {
		return 0, err
	}
	rr.Retry, off, err = unpackUint32(msg, off)
	if err != nil {
		return 0, err
	}
	rr.Expire, off, err = unpackUint32(msg, off)
	if err != nil {
		return 0, err
	}
	rr.MinTtl, off, err = unpackUint32(msg, off)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *MX) pack(msg []byte, off int, compression map[string]uint16) (off1 int, err error) {
	off, err = packUint16(rr.Preference, msg, off)
	if err != nil {
		return 0, err
	}

	off, err = packDomainName(rr.Exchange, msg, off, compression)
	if err != nil {
		return 0, err
	}

	return off, nil
}

func (rr *MX) unpack(msg []byte, off int) (off1 int, err error) {
	rr.Preference, off, err = unpackUint16(msg, off)
	if err != nil {
		return 0, err
	}

	rr.Exchange, off, err = unpackDomainName(msg, off)
	if err != nil {
		return 0, err
	}

	return off, nil
}
