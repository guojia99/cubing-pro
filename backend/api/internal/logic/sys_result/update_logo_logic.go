package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLogoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLogoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogoLogic {
	return &UpdateLogoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogoLogic) UpdateLogo(req *types.UpdateLogoReq) (resp *types.UpdateLogoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
