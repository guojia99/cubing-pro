package advertisements

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"gorm.io/gorm"
)

const (
	maxImageBytes        = 5 << 20
	maxImagePixels       = 20_000_000
	maxImageDimension    = 8192
	maxImageRequestBytes = (maxImageBytes+2)/3*4 + (64 << 10)
)

type ImageRequest struct {
	ImageName string `json:"ImageName"`
	Data      string `json:"Data"`
}

type ImageResponse struct {
	UID    string `json:"uid"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type validatedImage struct {
	data   []byte
	format string
	width  int
	height int
}

func decodeImage(data string) (validatedImage, error) {
	if len(data) == 0 || len(data) > base64.StdEncoding.EncodedLen(maxImageBytes) {
		return validatedImage{}, errors.New("image must contain at most 5 MiB")
	}
	raw, err := base64.StdEncoding.Strict().DecodeString(data)
	if err != nil || len(raw) > maxImageBytes {
		return validatedImage{}, errors.New("invalid base64 image or image exceeds 5 MiB")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png") {
		return validatedImage{}, errors.New("only valid jpeg and png images are allowed")
	}
	if config.Width < 1 || config.Height < 1 || config.Width > maxImageDimension || config.Height > maxImageDimension || int64(config.Width)*int64(config.Height) > maxImagePixels {
		return validatedImage{}, errors.New("image exceeds 8192 pixels per dimension or 20 million pixels")
	}
	_, _, err = image.Decode(bytes.NewReader(raw))
	if err != nil {
		return validatedImage{}, fmt.Errorf("decode image: %w", err)
	}
	// Preserve metadata such as EXIF orientation and PNG color information.
	return validatedImage{data: raw, format: format, width: config.Width, height: config.Height}, nil
}

func saveImage(db *gorm.DB, root string, userID uint, img validatedImage) (ImageResponse, error) {
	var out ImageResponse
	dir := filepath.Join(root, "advertisements")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return out, fmt.Errorf("create advertisement directory: %w", err)
	}
	uid := uuid.NewString()
	filePath := filepath.Join(dir, uid+"."+img.format)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return out, fmt.Errorf("create advertisement image: %w", err)
	}
	_, err = file.Write(img.data)
	err = errors.Join(err, file.Close())
	if err == nil {
		record := system.Image{UserID: userID, Use: "advertisement", Name: "广告图片", UID: uid, LocalPath: filePath}
		err = db.Create(&record).Error
	}
	if err != nil {
		return out, errors.Join(fmt.Errorf("save advertisement image: %w", err), os.Remove(filePath))
	}
	return ImageResponse{UID: uid, URL: "/v3/cube-api/static/image/" + uid, Width: img.width, Height: img.height}, nil
}

// UploadImage relies on the existing super-admin route group for authorization.
func UploadImage(db *gorm.DB, staticPath string, authUserID func(*gin.Context) (uint, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := authUserID(c)
		if err != nil {
			return
		}
		var req ImageRequest
		if err = decodeJSON(c, maxImageRequestBytes, &req); err != nil {
			requestError(c, err)
			return
		}
		img, err := decodeImage(req.Data)
		if err != nil {
			requestError(c, err)
			return
		}
		out, err := saveImage(db.WithContext(c.Request.Context()), staticPath, userID, img)
		if err != nil {
			exception.ErrInternalServer.ResponseWithError(c, "保存广告图片失败")
			return
		}
		exception.ResponseOK(c, out)
	}
}
