package initer

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/algs"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"

	"github.com/guojia99/go-tables/table"
)

func initDBCmd(svc **svc2.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "初始化数据库 event",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initEvent((*svc).DB); err != nil {
				return err
			}
			return nil
		},
	}

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

func formulasCmd(svc **svc2.Svc) *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "formulas",
		Short: "首次导入公式目录",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			service := *svc
			trainerPath := path
			if trainerPath == "" {
				trainerPath = service.Cfg.GlobalConfig.AlgTrainersPath
			}
			if trainerPath == "" {
				return fmt.Errorf("formula trainer path is required: use --path or set global.algTrainersPath")
			}
			if err := algs.InitCatalog(service.DB, trainerPath); err != nil {
				return err
			}
			var classes, cases int64
			if err := service.DB.Raw("SELECT (SELECT COUNT(*) FROM formula_classes), (SELECT COUNT(*) FROM formula_cases)").Row().Scan(&classes, &cases); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "公式目录已就绪：%d 个分类，%d 个案例。\n", classes, cases)
			return err
		},
	}
	cmd.Flags().StringVarP(&path, "path", "p", "", "Alg-Trainers 目录路径（相邻目录须有 Algs/catalog.json）")
	return cmd
}

func NewCmd(svc **svc2.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化数据库相关",
	}

	cmd.AddCommand(
		initDBCmd(svc),
		eventsListCmd(),
		formulasCmd(svc),
	)
	return cmd
}
