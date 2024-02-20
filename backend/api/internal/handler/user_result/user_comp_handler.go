package user_result

import (
	"net/http"

	"github.com/guojia99/cubing-pro/backend/api/internal/logic/user_result"
	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserCompHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserCompReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user_result.NewUserCompLogic(r.Context(), svcCtx)
		resp, err := l.UserComp(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
