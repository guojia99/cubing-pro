package user_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterCompAddEventLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterCompAddEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterCompAddEventLogic {
	return &RegisterCompAddEventLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterCompAddEventLogic) RegisterCompAddEvent(req *types.RegisterCompAddEventReq) (resp *types.RegisterCompAddEventResp, err error) {
	// todo: add your logic here and delete this line

	return
}
