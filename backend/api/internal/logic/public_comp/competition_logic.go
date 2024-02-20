package public_comp

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompetitionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompetitionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompetitionLogic {
	return &CompetitionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompetitionLogic) Competition(req *types.CompetitionReq) (resp *types.CompetitionResp, err error) {
	// todo: add your logic here and delete this line

	return
}
