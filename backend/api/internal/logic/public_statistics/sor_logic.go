package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SorLogic {
	return &SorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SorLogic) Sor(req *types.SorReq) (resp *types.SorResp, err error) {
	// todo: add your logic here and delete this line

	return
}
