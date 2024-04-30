package admin

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var config string
	var name string
	var password string

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "设置超级管理员",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := svc2.NewAPISvc(config)
			if err != nil {
				return err
			}

			user := user2.User{
				Model:    basemodel.Model{},
				Name:     name,
				EnName:   name,
				Password: password,
				LoginID:  name,
			}

			user.SetAuth(user2.AuthAdmin)
			user.SetAuth(user2.AuthSuperAdmin)
			user.SetAuth(user2.AuthDelegates)
			user.SetAuth(user2.AuthOrganizers)

			err = svc.DB.Save(&user).Error
			return err
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&config, "config", "c", "./etc/server.yaml", "配置文件")
	flags.StringVarP(&name, "name", "n", "admin", "超级管理员用户名")
	flags.StringVarP(&password, "password", "p", "admin", "超级管理员密码")

	return cmd
}
