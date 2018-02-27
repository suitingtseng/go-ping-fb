package main

import (
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	content = "\x5a\x94\x90\x00\x00\x01\x3e\xa1\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x20\x21\x22\x23\x24\x25\x26\x27\x28\x29\x2a\x2b\x2c\x2d\x2e\x2f\x30\x31\x32\x33\x34\x35\x36\x37"
	word    = 1 << 16
)

var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}

func main() {
	addrs, err := net.LookupHost("facebook.com")
	if err != nil {
		log.Fatalln(err.Error())
	}

	ip := net.ParseIP(addrs[0])
	var ct string
	var icmp_type byte
	if len([]byte(ip)) == 4 {
		ct = "ip4:icmp"
		icmp_type = '\x08'
	} else {
		ct = "ip6:ipv6-icmp"
		icmp_type = '\x80'
	}

	conn, err := net.Dial(ct, addrs[0])
	if err != nil {
		log.Fatalln(err.Error())
	}

	identifier, seq := r.Int()%word, 0

	b := make([]byte, 0)
	b = append(b, icmp_type, '\x00') // icmp type and code
	b = append(b, checksum(identifier, seq, content)...)
	b = append(b, byte(identifier/256), byte(identifier%256))
	b = append(b, byte(seq/256), byte(seq%256))
	b = append(b, []byte(content)...)

	n, err := conn.Write(b)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()

	log.Printf("PING facebook.com: %d data bytes\n", n)

	rbuf := make([]byte, 512)
	n, err = conn.Read(rbuf)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("PONG %v: %d data bytes\n", ip, n)
}

func checksum(identifier int, seq int, content string) []byte {
	s := identifier + seq + int('\x08')<<8
	cb := []byte(content)
	if odd(len(cb)) {
		cb = append(cb, '\x00')
	}
	for idx, b := range cb {
		if odd(idx) {
			s += int(b)
		} else {
			s += int(b) << 8
		}
	}
	for s >= word {
		d, r := s/word, s%word
		s = d + r
	}

	s = 1<<16 - 1 - s
	return []byte{byte(s / 256), byte(s % 256)}
}

func odd(i int) bool {
	return i%2 == 1
}
