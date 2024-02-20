package public_player

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayerNemesisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayerNemesisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerNemesisLogic {
	return &PlayerNemesisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayerNemesisLogic) PlayerNemesis(req *types.PlayerNemesisReq) (resp *types.PlayerNemesisResp, err error) {
	// todo: add your logic here and delete this line

	return
}
