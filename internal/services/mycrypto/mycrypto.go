package mycrypto

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const SecretKey = "8ha37nlpa4"

func HashFunc(str string) string {
	initString := fmt.Sprintf("%v:%v", str, SecretKey)

	return fmt.Sprintf("%x", md5.Sum([]byte(initString)))
}

func CreateRandomToken(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b) // записываем байты в слайс b
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}

	return hex.EncodeToString(b), nil
}
