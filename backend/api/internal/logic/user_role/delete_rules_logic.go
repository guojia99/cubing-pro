package user_role

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRulesLogic {
	return &DeleteRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRulesLogic) DeleteRules(req *types.DeleteRulesReq) (resp *types.DeleteRulesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
