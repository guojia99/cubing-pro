package robot

import (
	"fmt"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	robot2 "github.com/guojia99/cubing-pro/src/robot"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var config string

	cmd := &cobra.Command{
		Use:   "robot",
		Short: "魔方赛事网Robot",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := svc2.NewAPISvc(config)
			if err != nil {
				return err
			}
			robot := robot2.NewRobot(svc)
			fmt.Println("开始启动Robot")
			return robot.Run(cmd.Context())
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&config, "config", "c", "./etc/server.yaml", "配置文件")

	return cmd
}
