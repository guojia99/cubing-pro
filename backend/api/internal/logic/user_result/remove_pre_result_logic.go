package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemovePreResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemovePreResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePreResultLogic {
	return &RemovePreResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemovePreResultLogic) RemovePreResult(req *types.RemovePreResultReq) (resp *types.RemovePreResultResp, err error) {
	// todo: add your logic here and delete this line

	return
}
