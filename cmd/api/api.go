package api

import (
	"fmt"

	api2 "github.com/guojia99/cubing-pro/src/api"
	"github.com/guojia99/cubing-pro/src/internel/algs"
	"github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

func NewCmd(svc **svc2.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "魔方赛事网API, 单独开启API服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initFormulaService(*svc); err != nil {
				return err
			}
			api := api2.NewAPI(*svc)
			fmt.Println("开始启动API")
			return api.Run((*svc).Cfg.APIConfig.Host, (*svc).Cfg.APIConfig.Port)
		},
	}
	return cmd
}

func initFormulaService(service *svc2.Svc) error {
	if err := algs.InitCatalog(service.DB, service.Cfg.GlobalConfig.AlgTrainersPath); err != nil {
		return err
	}
	return formulaforum.Migrate(service.DB)
}
