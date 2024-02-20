package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserCompsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCompsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCompsLogic {
	return &UserCompsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCompsLogic) UserComps(req *types.UserCompsReq) (resp *types.UserCompsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
