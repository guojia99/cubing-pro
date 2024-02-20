package public_statistics

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MedalCollectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMedalCollectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MedalCollectionLogic {
	return &MedalCollectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MedalCollectionLogic) MedalCollection(req *types.MedalCollectionReq) (resp *types.MedalCollectionResp, err error) {
	// todo: add your logic here and delete this line

	return
}
