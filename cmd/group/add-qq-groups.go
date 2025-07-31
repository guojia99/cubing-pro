package group

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

type reqAddQQGroupsFlag struct {
	GroupID int `json:"group_id"`

	QQGroups             string `json:"qq_groups"`
	QQGroupUid           string `json:"qq_group_uid"`
	AddAssOrganizerUsers string `json:"add_ass_organizer_users"`
}

func UpdateQQGroups(svc **svc2.Svc) *cobra.Command {
	var req reqAddQQGroupsFlag
	cmd := &cobra.Command{
		Use:   "update-qq-group",
		Short: "添加比赛群组",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddQQGroup(*svc, req)
		},
	}
	flags := cmd.Flags()
	flags.IntVar(&req.GroupID, "group_id", 0, "group_id")
	flags.StringVar(&req.QQGroups, "qq_groups", "", "qq_groups")
	flags.StringVar(&req.QQGroupUid, "qq_group_uid", "", "qq_group_uid")
	flags.StringVar(&req.AddAssOrganizerUsers, "add_ass_organizer_users", "", "add_ass_organizer_users")
	return cmd
}

func runAddQQGroup(svc *svc2.Svc, req reqAddQQGroupsFlag) error {
	var cg competition.CompetitionGroup
	if err := svc.DB.Where("id = ?", req.GroupID).First(&cg).Error; err != nil {
		return err
	}

	if req.QQGroups != "" {
		cg.QQGroups = req.QQGroups
	}
	if req.QQGroupUid != "" {
		cg.QQGroupUid = req.QQGroupUid
	}
	if req.AddAssOrganizerUsers != "" {
		var og user.Organizers
		if err := svc.DB.Where("id = ?", cg.OrganizersID).Error; err != nil {
			return err
		}
		og.SetUsersCubingID([]string{req.AddAssOrganizerUsers})
		svc.DB.Save(&og)
	}

	return svc.DB.Save(&cg).Error
}
