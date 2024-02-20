package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddOrganizersPersonLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddOrganizersPersonLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOrganizersPersonLogic {
	return &AddOrganizersPersonLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddOrganizersPersonLogic) AddOrganizersPerson(req *types.AddOrganizersPersonReq) (resp *types.AddOrganizersPersonResp, err error) {
	// todo: add your logic here and delete this line

	return
}
