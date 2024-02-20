package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrganizerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizerLogic {
	return &GetOrganizerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizerLogic) GetOrganizer(req *types.GetOrganizerReq) (resp *types.GetOrganizerResp, err error) {
	// todo: add your logic here and delete this line

	return
}
