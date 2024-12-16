package robot

import (
	"fmt"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	robot2 "github.com/guojia99/cubing-pro/src/robot"
	"github.com/spf13/cobra"
)

func NewCmd(svc **svc2.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "robot",
		Short: "魔方赛事网Robot, 单独开启机器人服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			robot := robot2.NewRobot(*svc)
			fmt.Println("开始启动Robot")
			return robot.Run(cmd.Context())
		},
	}
	return cmd
}
