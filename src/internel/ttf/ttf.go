package ttf

import (
	_ "embed"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed HuaWenHeiTi.ttf
var huaWenHeiTiTTF []byte

func HuaWenHeiTiTTFFontFace(points float64) font.Face {
	f, _ := truetype.Parse(huaWenHeiTiTTF)
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
	})
	return face
}
