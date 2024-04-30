package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

// Encrypt 加密函数
func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建一个GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 创建一个随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 将时间戳编码成字符串并加密
	timestamp := time.Now().Unix()
	plaintext += fmt.Sprintf(":%d", timestamp)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func EncryptToByte(plaintext string, key []byte) ([]byte, error) {
	out, err := Encrypt(plaintext, key)
	return []byte(out), err
}

// Decrypt 解密函数
func Decrypt(ciphertext string, key []byte) (string, error) {
	// 先将密文解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 先解密
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], string(data[nonceSize:])
	plaintext, err := gcm.Open(nil, nonce, []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}

	// 分割出时间戳并返回
	parts := bytes.Split(plaintext, []byte(":"))
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid plaintext format")
	}

	return string(parts[0]), nil
}
