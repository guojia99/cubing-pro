package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOrganizersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteOrganizersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrganizersLogic {
	return &DeleteOrganizersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrganizersLogic) DeleteOrganizers(req *types.DeleteOrganizersReq) (resp *types.DeleteOrganizersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
