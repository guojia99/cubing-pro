package sys_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNameLogic {
	return &UpdateNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNameLogic) UpdateName(req *types.UpdateNameReq) (resp *types.UpdateNameResp, err error) {
	// todo: add your logic here and delete this line

	return
}
