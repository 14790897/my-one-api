package common

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

func Sha256Raw(data string) []byte {
	h := sha256.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func Sha1Raw(data []byte) []byte {
	h := sha1.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func Sha1(data string) string {
	return hex.EncodeToString(Sha1Raw([]byte(data)))
}

func HmacSha256Raw(message, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func HmacSha256(message, key string) string {
	return hex.EncodeToString(HmacSha256Raw([]byte(message), []byte(key)))
}

func RandomBytes(length int) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

func RandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	randomBytes := RandomBytes(length)
	for i := 0; i < length; i++ {
		result[i] = chars[randomBytes[i]%byte(len(chars))]
	}

	return string(result)
}

func RandomHex(length int) string {
	const chars = "abcdef0123456789"
	result := make([]byte, length)
	randomBytes := RandomBytes(length)
	for i := 0; i < length; i++ {
		result[i] = chars[randomBytes[i]%byte(len(chars))]
	}

	return string(result)
}

func RandomNumber(length int) string {
	const chars = "0123456789"
	result := make([]byte, length)
	randomBytes := RandomBytes(length)
	for i := 0; i < length; i++ {
		result[i] = chars[randomBytes[i]%byte(len(chars))]
	}

	return string(result)
}

func RandomUUID() string {
	all := RandomHex(32)
	return all[:8] + "-" + all[8:12] + "-" + all[12:16] + "-" + all[16:20] + "-" + all[20:]
}
