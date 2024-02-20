package organizers

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrganizersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrganizersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrganizersLogic {
	return &CreateOrganizersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrganizersLogic) CreateOrganizers(req *types.CreateOrganizersReq) (resp *types.CreateOrganizersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
