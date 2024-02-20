package user_role

import (
	"net/http"

	"github.com/guojia99/cubing-pro/backend/api/internal/logic/user_role"
	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UnbindRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnbindRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user_role.NewUnbindRoleLogic(r.Context(), svcCtx)
		resp, err := l.UnbindRole(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
