package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFooterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFooterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFooterLogic {
	return &UpdateFooterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFooterLogic) UpdateFooter(req *types.UpdateFooterReq) (resp *types.UpdateFooterResp, err error) {
	// todo: add your logic here and delete this line

	return
}
