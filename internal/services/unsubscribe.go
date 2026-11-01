package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

// ErrInvalidUnsubscribeToken is returned for malformed or forged unsubscribe tokens.
var ErrInvalidUnsubscribeToken = errors.New("This unsubscribe link is invalid. Please use the link from a recent email or contact support.")

func normalizeEmail(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}

func unsubscribeMAC(secret, addr string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("unsub:" + addr))
	return mac.Sum(nil)
}

// UnsubscribeToken returns a stateless, per-recipient token for the unsubscribe link:
// base64url(email) + "." + hex(HMAC-SHA256(secret, "unsub:"+email)).
func UnsubscribeToken(secret, addr string) string {
	addr = normalizeEmail(addr)
	return base64.RawURLEncoding.EncodeToString([]byte(addr)) + "." + hex.EncodeToString(unsubscribeMAC(secret, addr))
}

// ParseUnsubscribeToken verifies a token and returns the (normalized) email it was issued for.
func ParseUnsubscribeToken(secret, token string) (string, error) {
	encoded, sig, ok := strings.Cut(strings.TrimSpace(token), ".")
	if !ok {
		return "", ErrInvalidUnsubscribeToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalidUnsubscribeToken
	}
	gotMAC, err := hex.DecodeString(sig)
	if err != nil {
		return "", ErrInvalidUnsubscribeToken
	}
	addr := normalizeEmail(string(raw))
	if addr == "" || !hmac.Equal(gotMAC, unsubscribeMAC(secret, addr)) {
		return "", ErrInvalidUnsubscribeToken
	}
	return addr, nil
}
