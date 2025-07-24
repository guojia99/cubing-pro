package _interface

import (
	"fmt"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"

	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type UserI interface {
	GetCubeID(name string, createYear ...int) string
	MergeUser(baseCubeId string, otherCubeId string) error
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

// MergeUser 合并用户
func (c *UserIter) MergeUser(baseCubeId string, otherCubeId string) (err error) {
	// 查询用户
	var base user.User
	var other user.User
	if err = c.DB.First(&base, "cube_id = ?", baseCubeId).Error; err != nil {
		return err
	}
	if err = c.DB.First(&other, "cube_id = ?", otherCubeId).Error; err != nil {
		return err
	}

	// todo 讨论表等需要合并

	// 查出base的成绩
	var baseResults []result.Results
	if err = c.DB.Where("user_id = ?", base.ID).Find(&baseResults).Error; err != nil {
		return err
	}
	var otherResults []result.Results
	if err = c.DB.Where("user_id = ?", other.ID).Find(&otherResults).Error; err != nil {
		return err
	}

	var baseResultMap = make(map[string]result.Results)
	for _, r := range baseResults {
		key := fmt.Sprintf("%d_%s_%d", r.CompetitionID, r.EventID, r.RoundNumber)
		baseResultMap[key] = r
	}

	// 修改成绩
	var deleteResults []result.Results
	var updateResults []result.Results

	for _, r := range otherResults {
		key := fmt.Sprintf("%d_%s_%d", r.CompetitionID, r.EventID, r.RoundNumber)
		if _, ok := baseResultMap[key]; ok {
			deleteResults = append(deleteResults, r)
			continue
		}

		r.UserID = base.ID
		r.PersonName = base.Name
		r.CubeID = base.CubeID
		updateResults = append(updateResults, r)
	}

	fmt.Println("len updateResults => ", len(updateResults))
	fmt.Println("len deleteResults => ", len(deleteResults))
	fmt.Println("baseResult=>", updateResults)

	// 修改报名表
	var baseRegs []competition.Registration
	var otherRegs []competition.Registration
	var baseRegsMap = make(map[uint]competition.Registration)

	if err = c.DB.Where("user_id = ?", base.ID).Find(&baseRegs).Error; err != nil {
		return err
	}
	if err = c.DB.Where("user_id = ?", other.ID).Find(&otherRegs).Error; err != nil {
		return err
	}
	for _, reg := range baseRegs {
		baseRegsMap[reg.ID] = reg
	}
	var deleteRegs []competition.Registration
	var updateRegs []competition.Registration
	for _, reg := range otherRegs {
		if _, ok := baseRegsMap[reg.ID]; ok {
			deleteRegs = append(deleteRegs, reg)
		} else {
			reg.UserID = base.ID
			reg.UserName = base.Name
			// todo 修改报名项目
			updateRegs = append(updateRegs, reg)
		}
	}

	// 更新用户
	err1 := c.DB.Save(&updateResults).Error
	err2 := c.DB.Save(&updateRegs).Error

	// 删除
	var deleteRegIds []uint
	for _, reg := range deleteRegs {
		deleteRegIds = append(deleteRegIds, reg.ID)
	}
	var deleteResultIds []uint
	for _, r := range deleteResults {
		deleteResultIds = append(deleteResultIds, r.ID)
	}

	err3 := c.DB.Where("id IN ?", deleteRegIds).Delete(&competition.Registration{}).Error
	err4 := c.DB.Where("id IN ?", deleteResults).Delete(&result.Results{}).Error
	err5 := c.DB.Delete(&other).Error

	fmt.Println(err1, err2, err3, err4, err5)

	return nil
}
