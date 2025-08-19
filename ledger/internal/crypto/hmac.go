package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GitHubStyleMAC(secret, body []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(body)
	sum := m.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return "sha256=" + string(dst)
}

// VerifyGitHubSignature checks header value "sha256=<hex>" against body.
func VerifyGitHubSignature(secret, body []byte, headerVal string) bool {
	const p = "sha256="
	if len(headerVal) <= len(p) || headerVal[:len(p)] != p {
		return false
	}
	want := GitHubStyleMAC(secret, body)
	return hmac.Equal([]byte(headerVal), []byte(want))
}

// RawMACHex returns just the hex hash for our internal API HMAC header.
func RawMACHex(secret, body []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(body)
	sum := m.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return string(dst)
}
