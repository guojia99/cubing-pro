package public_comp

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompetitionResultsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompetitionResultsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompetitionResultsLogic {
	return &CompetitionResultsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompetitionResultsLogic) CompetitionResults(req *types.CompetitionResultsReq) (resp *types.CompetitionResultsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
