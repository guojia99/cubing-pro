package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TopNLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTopNLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TopNLogic {
	return &TopNLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TopNLogic) TopN(req *types.TopNReq) (resp *types.TopNResp, err error) {
	// todo: add your logic here and delete this line

	return
}
