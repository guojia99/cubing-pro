package admin

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

func NewCmd(svc **svc2.Svc) *cobra.Command {
	var name string
	var password string

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "设置超级管理员",
		RunE: func(cmd *cobra.Command, args []string) error {

			user := user2.User{
				Model:    basemodel.Model{},
				Name:     name,
				EnName:   name,
				Password: password,
				LoginID:  name,
			}

			user.SetAuth(user2.AuthPlayer)
			user.SetAuth(user2.AuthAdmin)
			user.SetAuth(user2.AuthSuperAdmin)
			user.SetAuth(user2.AuthDelegates)
			user.SetAuth(user2.AuthOrganizers)
			return (*svc).DB.Save(&user).Error
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&name, "name", "n", "admin", "超级管理员用户名")
	flags.StringVarP(&password, "password", "p", "admin", "超级管理员密码")

	return cmd
}

/*
(4)CLL:
AS5  R2 F R U2 R U' R' U2 F' R
L3   (U) R U2 R2 F R F' R U2 R'
L4   (U) R' U R' U2 R U' R' U R U' R2
T6   R' U R U2 R2 F R F' R
(6)EG1:
S2   (U2) F R2 U' R2 F U' F2 U' R
S3   (U') R' F R U2 R U' R2 F2 R F'
L3   R' U R2 U' R2 U' F R2 U' R'
L6   (U) R' U2 F R U2 R U' R2 F
T4   (U') R' U F R2 U' R2 U' F U' R
H2   F' U R U' R2 F2 R U' F
(11)EG2:
AS3  (U2) R' F R F' R U R B2 R2
AS4  R' U2 R U' R2 F' R U' F R
S3   R U' R' F R' F' R' F2 R2
S5   R' U R' F R2 U' F R' F'
T4   R2 F2 R U' F R' F' R U R

T5   (U) R' F2 R U' R' U R' F R U' R
U4   R2'F2 R U R U2' R2' F R F' R
L1   R2 B2 R2 F R' F' R U R U' R'
L2   (U) R2 B2 R' U R U' R' F R' F'
L5   F R' F' R U R U' R B2 R2

L6   F R U' R' U' R U R' F R2 B2

*/
