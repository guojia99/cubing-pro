package public

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayersLogic {
	return &PlayersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayersLogic) Players(req *types.PlayersReq) (resp *types.PlayersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
