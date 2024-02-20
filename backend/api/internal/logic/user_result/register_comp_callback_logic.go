package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterCompCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterCompCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterCompCallbackLogic {
	return &RegisterCompCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterCompCallbackLogic) RegisterCompCallback(req *types.RegisterCompCallbackReq) (resp *types.RegisterCompCallbackResp, err error) {
	// todo: add your logic here and delete this line

	return
}
