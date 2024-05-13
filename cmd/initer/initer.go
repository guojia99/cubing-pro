package initer

import (
	"fmt"

	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"

	"github.com/guojia99/go-tables/table"
)

func initDBCmd() *cobra.Command {
	var config string

	cmd := &cobra.Command{
		Use:   "event",
		Short: "初始化数据库 event",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := svc2.NewAPISvc(config)
			if err != nil {
				return err
			}

			if err = initEvent(svc.DB); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&config, "config", "c", "./etc/server.yaml", "配置文件")

	return cmd
}

func eventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event_list",
		Short: "初始化预设列表",
		RunE: func(cmd *cobra.Command, args []string) error {
			tb := table.DefaultSimpleTable(events)
			fmt.Println(tb)
			return nil
		},
	}
	return cmd
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化数据库相关",
	}

	cmd.AddCommand(
		initDBCmd(),
		eventsListCmd(),
	)
	return cmd
}
