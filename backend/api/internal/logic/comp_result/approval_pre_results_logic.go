package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApprovalPreResultsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApprovalPreResultsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApprovalPreResultsLogic {
	return &ApprovalPreResultsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApprovalPreResultsLogic) ApprovalPreResults(req *types.ApprovalPreResultsReq) (resp *types.ApprovalPreResultsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
