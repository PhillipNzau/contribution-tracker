package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func Encrypt(aesKey []byte, plaintext string) (string, error) {
	if len(aesKey) != 32 {
		return "", errors.New("aes key must be 32 bytes")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	pt := []byte(plaintext)
	ct := make([]byte, aes.BlockSize+len(pt))
	iv := ct[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ct[aes.BlockSize:], pt)
	return base64.URLEncoding.EncodeToString(ct), nil
}

func Decrypt(aesKey []byte, encoded string) (string, error) {
	if len(aesKey) != 32 {
		return "", errors.New("aes key must be 32 bytes")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	ct, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(ct) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ct[:aes.BlockSize]
	ct = ct[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ct, ct)
	return string(ct), nil
}