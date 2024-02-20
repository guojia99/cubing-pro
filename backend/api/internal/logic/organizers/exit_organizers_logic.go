package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExitOrganizersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExitOrganizersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExitOrganizersLogic {
	return &ExitOrganizersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExitOrganizersLogic) ExitOrganizers(req *types.ExitOrganizersReq) (resp *types.ExitOrganizersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
