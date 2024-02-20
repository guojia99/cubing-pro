package auth

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthCodeLogic {
	return &AuthCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthCodeLogic) AuthCode(req *types.AuthCodeReq) (resp *types.AuthCodeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
