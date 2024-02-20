package public_comp

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompetitionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompetitionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompetitionsLogic {
	return &CompetitionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompetitionsLogic) Competitions(req *types.CompetitionsReq) (resp *types.CompetitionsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
