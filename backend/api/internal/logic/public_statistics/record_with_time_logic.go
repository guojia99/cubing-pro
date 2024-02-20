package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordWithTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordWithTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordWithTimeLogic {
	return &RecordWithTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordWithTimeLogic) RecordWithTime(req *types.RecordWithTimeReq) (resp *types.RecordWithTimeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
