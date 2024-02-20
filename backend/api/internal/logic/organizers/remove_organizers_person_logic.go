package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveOrganizersPersonLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveOrganizersPersonLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveOrganizersPersonLogic {
	return &RemoveOrganizersPersonLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveOrganizersPersonLogic) RemoveOrganizersPerson(req *types.RemoveOrganizersPersonReq) (resp *types.RemoveOrganizersPersonResp, err error) {
	// todo: add your logic here and delete this line

	return
}
