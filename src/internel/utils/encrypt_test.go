package utils

import (
	"fmt"
	"testing"
)

func TestEncrypt(t *testing.T) {
	key := []byte("cubing-pro123456")

	plaintext := "cubing-pro-text"
	// 加密
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}
	fmt.Println("Encrypted:", encrypted)

	// 解密
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		fmt.Println("Decryption error:", err)
		return
	}
	fmt.Println("Decrypted:", decrypted)
}
