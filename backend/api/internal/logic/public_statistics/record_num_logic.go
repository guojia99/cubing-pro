package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordNumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordNumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordNumLogic {
	return &RecordNumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordNumLogic) RecordNum(req *types.RecordNumReq) (resp *types.RecordNumResp, err error) {
	// todo: add your logic here and delete this line

	return
}
