package public_player

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayerRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayerRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerRecordsLogic {
	return &PlayerRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayerRecordsLogic) PlayerRecords(req *types.PlayerRecordsReq) (resp *types.PlayerRecordsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
