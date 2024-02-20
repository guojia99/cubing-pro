package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCompLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCompLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompLogic {
	return &UpdateCompLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCompLogic) UpdateComp(req *types.UpdateCompReq) (resp *types.UpdateCompResp, err error) {
	// todo: add your logic here and delete this line

	return
}
