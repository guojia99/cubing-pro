package system

import basemodel "github.com/guojia99/cubing-pro/backend/pkg/model/base"

const MaxImageSize = 1024 * 1024 * 3 / 2 // base64 size 1M

// Image 保存一些系统图片和events图片的东西
type Image struct {
	basemodel.DBModel

	Name  string `gorm:"column:name"`
	Image string `gorm:"column:image"` // base64
}
