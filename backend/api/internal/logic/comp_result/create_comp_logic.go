package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCompLogic {
	return &CreateCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCompLogic) CreateComp(req *types.CreateCompReq) (resp *types.CreateCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
