package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	n = 32768
	r = 8
	p = 1
	s = 16
	k = 32
)

func Hash(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	dk, err := scrypt.Key([]byte(password), salt, n, r, p, k)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d$%d$%d$%s$%s", n, r, p, encodeBase64(salt), encodeBase64(dk)), nil
}

func Compare(hashed, password string) bool {
	n, r, p, salt, dk := decode(hashed)
	if len(dk) == 0 {
		return false
	}

	pk, err := scrypt.Key([]byte(password), salt, n, r, p, len(dk))
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(dk, pk) == 1
}

func decode(hashed string) (n, r, p int, salt, dk []byte) {
	xs := strings.Split(hashed, "$")
	if len(xs) != 5 {
		return
	}

	var err error
	n, err = strconv.Atoi(xs[0])
	if err != nil {
		return
	}
	r, err = strconv.Atoi(xs[1])
	if err != nil {
		return
	}
	p, err = strconv.Atoi(xs[2])
	if err != nil {
		return
	}
	salt = decodeBase64(xs[3])
	dk = decodeBase64(xs[4])
	return
}

func generateSalt() ([]byte, error) {
	b := make([]byte, s)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func encodeBase64(p []byte) string {
	return base64.RawStdEncoding.EncodeToString(p)
}

func decodeBase64(s string) []byte {
	p, _ := base64.RawStdEncoding.DecodeString(s)
	return p
}
