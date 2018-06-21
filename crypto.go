package ptk

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"

	"golang.org/x/crypto/bcrypt"
)

const DefaultBcryptRounds = 12

var b64 = base64.RawURLEncoding

func HashPassword(password string, cost int) (string, error) {
	if cost == -1 {
		cost = DefaultBcryptRounds
	}

	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(h), err
}

func CheckPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func RandomSafeString(ln int) string {
	buf := make([]byte, ln)

	if _, err := rand.Read(buf); err != nil {
		log.Panic(err) // this is a panic because if it happens, something is extremely wrong with the server
	}

	return b64.EncodeToString(buf)
}

func CreateMAC(password, token, salt string) string {
	// if we change the token size to be > 16 bytes, we'll have to decode the token/salt otherwise they will get hashed
	h := hmac.New(sha256.New, []byte(token+salt))
	h.Write([]byte(password))
	return b64.EncodeToString(h.Sum(nil))
}

func VerifyMac(mac1, password, token, salt string) bool {
	mac2 := DecodeBase64(CreateMAC(password, token, salt))
	return hmac.Equal(DecodeBase64(mac1), mac2)
}

func DecodeBase64(s string) []byte {
	b, err := b64.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}
