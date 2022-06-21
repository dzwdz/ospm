package otp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"time"
)

type Ratelimited interface {
	Verify(p int) (bool, string)
	Peek(offset int) int // meant for testing only
}

type ratelimited struct {
	secret   []byte
	lastStep int64
}

func InitRatelimited(s string) Ratelimited {
	secret, err := DecodeSecret(s)
	if err != nil {
		panic("malformed secret")
	}

	return &ratelimited{
		secret: secret,
	}
}

func (self *ratelimited) Verify(p int) (bool, string) {
	// TODO ratelimit failed attempts
	baseStep := CurrentStep()
	for skew := -1; skew <= 1; skew++ {
		step := baseStep + int64(skew)
		if p == Current(step, self.secret) {
			// TODO make this atomic, prevent race conditions
			if step <= self.lastStep {
				return false, "ratelimited.success"
			}
			self.lastStep = step
			return true, ""
		}
	}
	return false, "invalid"
}

func (self *ratelimited) Peek(offset int) int {
	return Current(CurrentStep()+int64(offset), self.secret)
}

func DecodeSecret(s string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(s)
}

func CurrentStep() int64 {
	return time.Now().Unix() / 30
}

func Current(step int64, secret []byte) int {
	mac := hmac.New(sha1.New, secret)
	binary.Write(mac, binary.BigEndian, step)

	sum := mac.Sum(nil)

	offset := int(sum[19] & 0xf)
	bin_code := (int(sum[offset+0]) & 0x7f) << 24
	bin_code |= (int(sum[offset+1]) & 0xff) << 16
	bin_code |= (int(sum[offset+2]) & 0xff) << 8
	bin_code |= (int(sum[offset+3]) & 0xff)

	return bin_code % 1_000_000
}
