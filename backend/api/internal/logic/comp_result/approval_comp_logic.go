package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApprovalCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApprovalCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApprovalCompLogic {
	return &ApprovalCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApprovalCompLogic) ApprovalComp(req *types.ApprovalCompReq) (resp *types.ApprovalCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
