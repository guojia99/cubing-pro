package api

import (
	"fmt"

	api2 "github.com/guojia99/cubing-pro/src/api"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

func NewCmd(svc **svc2.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "魔方赛事网API, 单独开启API服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := api2.NewAPI(*svc)
			if err != nil {
				return err
			}
			fmt.Println("开始启动API")
			return api.Run((*svc).Cfg.APIConfig.Host, (*svc).Cfg.APIConfig.Port)
		},
	}
	return cmd
}
