package user_role

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddRulesLogic {
	return &AddRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddRulesLogic) AddRules(req *types.AddRulesReq) (resp *types.AddRulesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
