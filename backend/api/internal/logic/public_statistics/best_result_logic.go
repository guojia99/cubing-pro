package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BestResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBestResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BestResultLogic {
	return &BestResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BestResultLogic) BestResult(req *types.BestResultReq) (resp *types.BestResultResp, err error) {
	// todo: add your logic here and delete this line

	return
}
