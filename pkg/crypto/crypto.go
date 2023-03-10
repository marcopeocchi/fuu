package pkg

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"fuu/v/pkg/config"
	"math/rand"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/credentials"
)

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

// ********************************
// LEGACY HMAC BASED AUTHENTCATION
// IT WILL BE REMOVED
// ********************************

// Returns a Base64 encoded HMAC of a messagege given a secret
func computeHMAC(data, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Generate an expire date from 30 days to time.Now, then encrypt it
// with AES-128/192/256 according to len(secret).
func generateExpireDate(secret []byte) ([]byte, error) {
	cbc, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	t := time.Now().Add(time.Hour * 24 * 30).Format(time.RFC3339)
	src := []byte(t)
	dest := make([]byte, len(src))

	cbc.Encrypt(dest, src)
	src = nil

	return dest, nil
}

// Decrypt the expire date and verify if it is 30 days or more old
func validateExpireDate(message string, secret []byte) bool {
	cbc, err := aes.NewCipher(secret)
	if err != nil {
		return false
	}
	src, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return false
	}
	dest := make([]byte, len(src))

	cbc.Decrypt(dest, src)

	//lint:ignore SA1002 the layout is actually correct
	expireDate, err := time.Parse("2018-12-10T15:00", string(dest))
	if err != nil {
		return false
	}

	return time.Now().UnixMilli() < expireDate.UnixMilli()
}

// Generate a token encoded which parts are Base64 encoded.
// The token should be 118 bytes in lenght and have the following format:
//
// [Base64(id-token)].[Base64(AES-128(RFC 3339 date))].[Base64(thumbprint)]
//
// id-token: derived from user pass + server secret
//
// expireDate: RFC 3339 compliant (and ISO 8601) date encrypted with aes
// theoretically it shouldn't be need encrypt it.
//
// thumbprint: HMAC-256 of id-token+expireDate, provides
// message integrity and authenticity.
func GenerateTokenString(message, secret, chiperSecret []byte) (string, error) {
	token := computeHMAC(message, secret)
	expireDate, err := generateExpireDate(chiperSecret)
	if err != nil {
		return "", err
	}
	buff := append(expireDate, token...)
	thumbprint := computeHMAC(buff, secret)

	expireDateString := base64.StdEncoding.EncodeToString(expireDate)

	return fmt.Sprintf("%s.%s.%s", token, expireDateString, thumbprint), nil
}

// Recompute the token and perform a validation against the given one.
func ValidateTokenString(token string, message, secret []byte) (bool, error) {
	// token[0] expiration[1] thumbprint[2]
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, errors.New("token malformed")
	}

	tokenToMatch, err := GenerateTokenString(message, secret, secret)
	if err != nil {
		return false, err
	}

	matches := strings.Split(tokenToMatch, ".")

	// token expiration
	if validateExpireDate(parts[1], secret) {
		return false, errors.New("token expired")
	}
	// thumbprint verification
	if parts[2] != matches[2] {
		return false, errors.New("thumbprint does not match")
	}
	// token verification
	if parts[0] != matches[0] {
		return false, errors.New("invalid")
	}

	return true, nil
}

// Generate a token encoded which parts are Base64 encoded.
// The token should be 118 bytes in lenght and have the following format:
//
// [Base64(id-token)].[Base64(AES-128(RFC 3339 date))].[Base64(thumbprint)]
//
// id-token: derived from user pass + server secret
//
// expireDate: RFC 3339 compliant (and ISO 8601) date encrypted with aes
// theoretically it shouldn't be need encrypt it.
//
// thumbprint: HMAC-256 of id-token+expireDate, provides
// message integrity and authenticity.
func GenerateToken2String(message, secret []byte) (string, error) {
	token := computeHMAC(message, secret)
	thumbprint := computeHMAC([]byte(token), secret)

	return fmt.Sprintf("%s.%s", token, thumbprint), nil
}

func GenerateRandomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)

	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache = rand.Int63()
			remain = letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func LoadTLSCreds() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(config.Instance().TLSCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
