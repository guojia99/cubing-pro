package root

import (
	"github.com/donnie4w/go-logger/logger"
	"github.com/guojia99/cubing-pro/cmd/admin"
	"github.com/guojia99/cubing-pro/cmd/gateway"
	"github.com/guojia99/cubing-pro/cmd/group"
	"github.com/guojia99/cubing-pro/cmd/initer"
	"github.com/guojia99/cubing-pro/cmd/robot"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"

	"github.com/guojia99/cubing-pro/cmd/api"
)

func NewRootCmd() *cobra.Command {
	var s *svc.Svc
	var config string
	var runJob bool
	var runScramble bool = true
	cmd := &cobra.Command{
		Use:   "cubing-pro",
		Short: "魔方赛事网",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			s, err = svc.NewAPISvc(config, runJob, runScramble)
			if err != nil {
				return err
			}
			_, _ = logger.SetRollingFile(s.Cfg.Log.Path, "cubing-pro.log", int64(s.Cfg.Log.MaxSize), logger.MB)
			logger.Infof("开始运行Cubing Pro...")
			return err
		},
	}
	flags := cmd.PersistentFlags()
	flags.StringVarP(&config, "config", "c", "./etc/server.yaml", "配置文件")
	flags.BoolVarP(&runJob, "job", "j", false, "运行定时任务")
	//flags.BoolVarP(&runScramble, "scramble", "s", false, "运行打乱组件")
	cmd.AddCommand(
		api.NewCmd(&s),
		robot.NewCmd(&s),
		admin.NewCmd(&s),
		initer.NewCmd(&s),
		gateway.NewCmd(&s),
		group.AddGroupNewCmd(&s),
		group.UpdateQQGroups(&s),
	)
	return cmd
}
