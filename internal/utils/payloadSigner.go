package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type PayloadSigner struct{}

func (s *PayloadSigner) Sign(secret []byte, data []byte) (signedData string) {
	mac := hmac.New(sha256.New, secret)
	mac.Write(data)

	sig := mac.Sum(nil)

	payload := base64.RawURLEncoding.EncodeToString(data)
	signature := base64.RawURLEncoding.EncodeToString(sig)

	return payload + "." + signature
}

func (s *PayloadSigner) Verify(secret []byte, value string) (data []byte, ok bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, false
	}

	payloadB64 := parts[0]
	sigB64 := parts[1]

	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, false
	}

	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(sig, expected) {
		return nil, false
	}

	return payload, true
}
