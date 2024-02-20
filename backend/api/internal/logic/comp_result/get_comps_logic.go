package comp_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompsLogic {
	return &GetCompsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompsLogic) GetComps(req *types.GetCompsReq) (resp *types.GetCompsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
