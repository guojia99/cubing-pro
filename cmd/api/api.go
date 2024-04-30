package api

import (
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"

	api2 "github.com/guojia99/cubing-pro/src/api"
)

func NewCmd() *cobra.Command {
	var config string

	cmd := &cobra.Command{
		Use:   "api",
		Short: "魔方赛事网API",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := svc2.NewAPISvc(config)
			if err != nil {
				return err
			}
			api := api2.NewAPI(svc)
			return api.Run(svc.Cfg.APIGatewayConfig.Host, svc.Cfg.APIGatewayConfig.APIPort)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&config, "config", "c", "./etc/server.yaml", "配置文件")

	return cmd
}
