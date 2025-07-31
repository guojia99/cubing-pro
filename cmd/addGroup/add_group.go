package addGroup

import (
	"errors"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

type reqFlag struct {
	LeaderCubeID string `json:"leader_cube_id"` // 主办的ID

	OrganizerName  string `json:"organizer_name"`
	OrganizerEmail string `json:"organizer_email"`

	GroupName  string `json:"group_name"`
	QQGroups   string `json:"qq_groups"`
	QQGroupUid string `json:"qq_group_uid"`
}

func NewCmd(svc **svc2.Svc) *cobra.Command {
	var req reqFlag
	cmd := &cobra.Command{
		Use:   "add-group",
		Short: "添加比赛群组",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddGroup(*svc, req)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&req.LeaderCubeID, "leader_cube_id", "", "CubeID")
	flags.StringVar(&req.OrganizerName, "organizer_name", "", "主办团队")
	flags.StringVar(&req.OrganizerEmail, "organizer_email", "", "主办团队邮箱")
	flags.StringVar(&req.GroupName, "group_name", "", "群组名")
	flags.StringVar(&req.QQGroups, "qq_groups", "", "QQ群")
	flags.StringVar(&req.QQGroupUid, "qq_group_uid", "", "qq官方机器人UID")

	// go run main.go  add-group -c ./local/server_local_dev.yaml --leader_cube_id 2024CUBE01  --organizer_email 2225470188@qq.com --organizer_name 星辰的魔方小窝 --group_name 星辰的魔方小窝 --qq_groups 428243498 --qq_group_uid 966F739CD99E3B4B3437BCB738400CB9
	return cmd
}

func runAddGroup(svc *svc2.Svc, req reqFlag) error {

	var cuber user.User
	if svc.DB.Where("cube_id = ?", req.LeaderCubeID).First(&cuber).Error != nil {
		return errors.New("查不到对应的cubeId")
	}

	cuber.SetAuth(user.AuthOrganizers)
	svc.DB.Save(&cuber)

	var og = user.Organizers{
		Name:         req.OrganizerName,
		Introduction: req.OrganizerName,
		Email:        req.OrganizerEmail,
		LeaderID:     cuber.CubeID,
		Status:       user.Using,
	}
	og.SetUsersCubingID([]string{"2023JIAY01"})
	if err := svc.DB.Save(&og).Error; err != nil {
		return err
	}

	var cg = competition.CompetitionGroup{
		Name:         req.GroupName,
		OrganizersID: og.ID,
		QQGroups:     req.QQGroups,
		QQGroupUid:   req.QQGroupUid,
	}
	svc.DB.Save(&cg)
	return nil
}
