package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrganizersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizersLogic {
	return &GetOrganizersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizersLogic) GetOrganizers(req *types.GetOrganizersReq) (resp *types.GetOrganizersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
