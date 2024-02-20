package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetireCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetireCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetireCompLogic {
	return &RetireCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetireCompLogic) RetireComp(req *types.RetireCompReq) (resp *types.RetireCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
