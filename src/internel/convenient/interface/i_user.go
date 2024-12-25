package _interface

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type UserI interface {
	GetCubeID(name string, createYear ...int) string
}

type UserIter struct {
	DB *gorm.DB
}

func (c *UserIter) GetCubeID(name string, createYear ...int) string {
	y := time.Now().Year()
	if len(createYear) > 0 {
		y = createYear[0]
	}
	baseName := utils.GetIDButNotNumber(name, y)

	var find []user.User
	c.DB.Model(&user.User{}).Where("cube_id like ?", fmt.Sprintf("%%%s%%", baseName)).Find(&find)
	num := len(find) + 1
	return fmt.Sprintf("%s%02d", baseName, num)
}
