package user_role

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRulesLogic {
	return &UpdateRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRulesLogic) UpdateRules(req *types.UpdateRulesReq) (resp *types.UpdateRulesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
