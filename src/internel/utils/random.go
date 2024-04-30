package utils

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/mojocn/base64Captcha"
)

func GenerateRandomKey(timestamp int64) []byte {
	var data = []byte(fmt.Sprintf("%dcubing-pro-key", timestamp))
	for len(data) < 32 {
		data = append(data, '=')
	}
	return data[:32]
}

func RandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// MathRandomConfig 生成图形化算术验证码配置
func MathRandomConfig() *base64Captcha.DriverMath {
	mathType := &base64Captcha.DriverMath{
		Height:          200,
		Width:           300,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		BgColor: &color.RGBA{
			R: 40,
			G: 30,
			B: 89,
			A: 29,
		},
		Fonts: nil,
	}
	return mathType
}
