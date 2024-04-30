package main

import (
	"github.com/guojia99/cubing-pro/cmd/admin"
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
		admin.NewCmd(),
	)
	return cmd
}

func main() {
	cmd := NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
