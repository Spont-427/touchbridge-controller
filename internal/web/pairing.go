package web

import (
	"crypto/rand"
	"math/big"
	"time"
)

// PairingTTL is the lifetime of a six-digit pairing PIN.
const PairingTTL = 10 * time.Minute

// GeneratePIN returns a uniformly selected six-digit PIN, including leading zeros.
func GeneratePIN() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmtSixDigits(value.Int64()), nil
}

func fmtSixDigits(value int64) string {
	// Avoid fmt.Sprintf in the pairing hot path while preserving leading zeros.
	var out [6]byte
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = byte('0' + value%10)
		value /= 10
	}
	return string(out[:])
}
