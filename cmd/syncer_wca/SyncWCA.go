package syncer_wca

import (
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

func NewStaticWCACmd(svc **svc.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync_wca",
		Short: "同步WCA数据库统计",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := *svc
			err := s.Wca.SyncStatic()
			return err
		},
	}
	return cmd
}
