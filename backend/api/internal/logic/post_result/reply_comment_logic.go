package post_result

import (
	"context"

	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReplyCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplyCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyCommentLogic {
	return &ReplyCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReplyCommentLogic) ReplyComment(req *types.ReplyCommentReq) (resp *types.ReplyCommentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
