package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path"
)

type UpdateAvatarReq struct {
	URL       string `json:"URL"`
	ImageName string `json:"ImageName"` // baseFileName
	Data      string `json:"Data"`      // base64
}

func UpdateAvatar(svc *svc.Svc) gin.HandlerFunc {
	avatarPath := path.Join(svc.Cfg.APIGatewayConfig.StaticPath, "avatar")
	_ = os.MkdirAll(avatarPath, os.ModePerm)

	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req UpdateAvatarReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		imgDB := system.Image{
			UserID:                  user.ID,
			Use:                     "avatar",
			Name:                    "头像",
			UID:                     uuid.NewString(),
			URL:                     req.URL,
			LocalPath:               "",
			LocalPathThumbnailImage: "",
		}

		if len(req.URL) > 0 {
			imgDB.Name = path.Base(req.URL)
		}

		if len(req.Data) > 0 {
			imgDB.Name = req.ImageName

			format := path.Ext(req.ImageName)
			//fmt.Println(req.Data)

			data, err := base64.StdEncoding.DecodeString(req.Data)
			if err != nil {
				exception.ErrValidationFailed.ResponseWithError(ctx, "图像解码错误")
				return
			}
			// 判断解码后的数据是否为图像
			img, format, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				exception.ErrValidationFailed.ResponseWithError(ctx, "不是有效的图像数据")
				return
			}

			// 保存到本地文件
			baseImagePath := path.Join(avatarPath, fmt.Sprintf("%s_{%s}.%s", imgDB.UID, req.ImageName, format))
			if err = saveImageToPath(img, baseImagePath, format); err != nil {
				exception.ErrValidationFailed.ResponseWithError(ctx, "保存原始格式数据失败")
				return
			}
			imgDB.LocalPath = baseImagePath

			// 缩略图
			maxWidth, maxHeight := 200, 200
			originalWidth := img.Bounds().Dx()
			originalHeight := img.Bounds().Dy()
			// 如果图像的宽度和高度都小于阈值，不需要缩放
			if !(originalWidth <= maxWidth || originalHeight <= maxHeight) {
				// 根据最大尺寸和原始尺寸的比例，计算缩放比例
				var newWidth, newHeight uint
				if originalWidth > originalHeight {
					newWidth = uint(maxWidth)
					newHeight = uint(float64(originalHeight) * float64(maxWidth) / float64(originalWidth))
				} else {
					newHeight = uint(maxHeight)
					newWidth = uint(float64(originalWidth) * float64(maxHeight) / float64(originalHeight))
				}
				thumb := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

				thumbImagePath := path.Join(avatarPath, fmt.Sprintf("%s_{%s}_thumb.%s", imgDB.UID, req.ImageName, format))
				if err = saveImageToPath(thumb, thumbImagePath, format); err != nil {
					exception.ErrValidationFailed.ResponseWithError(ctx, "保存缩略格式数据失败")
					return
				}
				imgDB.LocalPathThumbnailImage = thumbImagePath
			}
		}

		svc.DB.Save(&imgDB)

		user.Avatar = fmt.Sprintf("/static/image/%s", imgDB.UID)
		svc.DB.Save(&user)

		exception.ResponseOK(ctx, nil)
	}
}

func saveImageToPath(img image.Image, imgPath string, format string) error {
	baseImagePathOutPutFile, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	defer baseImagePathOutPutFile.Close()

	switch format {
	case "jpeg", ".jpeg", "jpg", ".jpg":
		err = jpeg.Encode(baseImagePathOutPutFile, img, &jpeg.Options{Quality: 65})
	case "png", ".png":
		err = png.Encode(baseImagePathOutPutFile, img)
	}
	return err
}
