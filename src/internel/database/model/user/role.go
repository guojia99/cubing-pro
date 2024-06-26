package user

//
//// 设计参考： https://www.cnblogs.com/wlovet/p/17717905.html
//
//import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
//
//// Role 角色表
//type Role struct {
//	basemodel.Model
//
//	Name       string `gorm:"column:name"`        // 角色名
//	CreateId   uint   `gorm:"column:create_id"`   // 创建角色人ID
//	ModifierId uint   `gorm:"column:modifier_id"` // 变更人ID
//}
//
//type AssUsersRoles struct {
//	basemodel.Model
//
//	UserId uint `gorm:"index:,unique,composite:AssUsersRoles"`
//	RoleId uint `gorm:"index:,unique,composite:AssUsersRoles"`
//}
//
//// AuthRule 权限表
//type AuthRule struct {
//	basemodel.Model
//
//	Code       uint   `gorm:"column:code"`        // 权限码
//	Name       string `gorm:"column:name"`        // 权限名
//	CreateId   uint   `gorm:"column:create_id"`   // 创建角色人ID
//	ModifierId uint   `gorm:"column:modifier_id"` // 变更人ID
//
//	Url    string `gorm:"column:url"`    // 权限生效路由 * 代表所有
//	Option string `gorm:"column:option"` // 权限可用的操作 * 代表所有
//}
//
//type AssRoleAuthRule struct {
//	basemodel.Model
//
//	RoleId     uint `gorm:"index:,unique,composite:AssRoleAuthRule"`
//	AuthRuleId uint `gorm:"index:,unique,composite:AssRoleAuthRule"`
//}
