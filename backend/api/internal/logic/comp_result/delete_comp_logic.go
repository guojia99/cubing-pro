package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCompLogic {
	return &DeleteCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCompLogic) DeleteComp(req *types.DeleteCompReq) (resp *types.DeleteCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
