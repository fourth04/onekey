package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

// FormatDate 格式化时间
func FormatDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// Encrypt is encrypt the data with salt
func Encrypt(data string, salt string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(salt))
	if err != nil {
		return "", err
	}
	cipher := hash.Sum(nil)

	buf := new(bytes.Buffer)
	buf.Write(cipher)
	buf.WriteString(data)
	_, err = hash.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// RandomString generate random string
func RandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
