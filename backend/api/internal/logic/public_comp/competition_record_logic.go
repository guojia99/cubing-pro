package public_comp

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompetitionRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompetitionRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompetitionRecordLogic {
	return &CompetitionRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompetitionRecordLogic) CompetitionRecord(req *types.CompetitionRecordReq) (resp *types.CompetitionRecordResp, err error) {
	// todo: add your logic here and delete this line

	return
}
