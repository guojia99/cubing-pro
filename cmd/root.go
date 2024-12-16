package root

import (
	"github.com/donnie4w/go-logger/logger"
	"github.com/guojia99/cubing-pro/cmd/admin"
	"github.com/guojia99/cubing-pro/cmd/gateway"
	"github.com/guojia99/cubing-pro/cmd/initer"
	"github.com/guojia99/cubing-pro/cmd/robot"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"

	"github.com/guojia99/cubing-pro/cmd/api"
)

func NewRootCmd() *cobra.Command {
	var s *svc.Svc
	var config string
	cmd := &cobra.Command{
		Use:   "cubing-pro",
		Short: "魔方赛事网",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			s, err = svc.NewAPISvc(config)
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
	cmd.AddCommand(
		api.NewCmd(&s),
		robot.NewCmd(&s),
		admin.NewCmd(&s),
		initer.NewCmd(&s),
		gateway.NewCmd(s),
	)
	return cmd
}
