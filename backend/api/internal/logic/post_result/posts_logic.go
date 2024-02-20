package post_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PostsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostsLogic {
	return &PostsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PostsLogic) Posts(req *types.PostsReq) (resp *types.PostsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
