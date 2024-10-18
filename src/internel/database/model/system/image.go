package system

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

//const MaxImageSize = 1024 * 1024 * 3 / 2 // base64 size 1M

// Image 保存一些系统图片和events图片的东西
type Image struct {
	basemodel.Model

	// 业务用途
	UserID uint   `gorm:"column:user_id"` // 用户ID
	Use    string `gorm:"column:d_use"`   // 用途
	Name   string `gorm:"column:name"`    // 图像名
	UID    string `gorm:"column:uid"`     // uuid4

	// 文件实际存储数据
	URL                     string `gorm:"column:url"`           // 链接
	LocalPath               string `gorm:"column:local_path"`    // 本地存储位置
	LocalPathThumbnailImage string `gorm:"column:local_path_ti"` // 本地缩略图
}
