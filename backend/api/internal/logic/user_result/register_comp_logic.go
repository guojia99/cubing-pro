package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterCompLogic {
	return &RegisterCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterCompLogic) RegisterComp(req *types.RegisterCompReq) (resp *types.RegisterCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
