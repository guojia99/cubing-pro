package database

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type userI interface {
	GetCubeID(name string) string
}

func (c *convenient) GetCubeID(name string) string {
	baseName := utils.GetIDButNotNumber(name)

	var find []user.User
	c.db.Model(&user.User{}).Find(&find, "%s", fmt.Sprintf("%%%s%%", baseName))

	num := len(find) + 1
	return fmt.Sprintf("%s%02d", baseName, num)
}
