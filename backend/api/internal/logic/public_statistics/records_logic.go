package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordsLogic {
	return &RecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordsLogic) Records(req *types.RecordsReq) (resp *types.RecordsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
