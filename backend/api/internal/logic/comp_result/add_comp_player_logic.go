package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddCompPlayerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCompPlayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCompPlayerLogic {
	return &AddCompPlayerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddCompPlayerLogic) AddCompPlayer(req *types.AddCompPlayerReq) (resp *types.AddCompPlayerResp, err error) {
	// todo: add your logic here and delete this line

	return
}
