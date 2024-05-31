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

func TestEncrypt1(t *testing.T) {
	key := GenerateRandomKey(12345)

	data, err := Encrypt("12345", key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestDecrypt(t *testing.T) {
	var data = "Y7bU2S+MX5B93UXbBIWb5Y2RfzF7Q/y/vYBAUXM8CNut4hHcGQ=="
	fmt.Println(string(GenerateRandomKey(12345)))
	value, err := Decrypt(data, GenerateRandomKey(12345))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(value)
}
