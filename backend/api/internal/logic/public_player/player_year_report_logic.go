package public_player

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlayerYearReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlayerYearReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlayerYearReportLogic {
	return &PlayerYearReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlayerYearReportLogic) PlayerYearReport(req *types.PlayerYearReportReq) (resp *types.PlayerYearReportResp, err error) {
	// todo: add your logic here and delete this line

	return
}
