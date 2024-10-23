package root

import (
	"github.com/guojia99/cubing-pro/cmd/admin"
	"github.com/guojia99/cubing-pro/cmd/initer"
	"github.com/guojia99/cubing-pro/cmd/robot"
	"github.com/spf13/cobra"

	"github.com/guojia99/cubing-pro/cmd/api"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cube-pro",
		Short: "魔方赛事网",
	}

	cmd.AddCommand(
		api.NewCmd(),
		robot.NewCmd(),
		admin.NewCmd(),
		initer.NewCmd(),
	)
	return cmd
}
