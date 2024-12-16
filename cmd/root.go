package root

import (
	"github.com/donnie4w/go-logger/logger"
	"github.com/guojia99/cubing-pro/cmd/admin"
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
	)
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
