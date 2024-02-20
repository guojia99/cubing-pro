package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApprovalCompsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApprovalCompsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApprovalCompsLogic {
	return &ApprovalCompsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApprovalCompsLogic) ApprovalComps(req *types.ApprovalCompsReq) (resp *types.ApprovalCompsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
