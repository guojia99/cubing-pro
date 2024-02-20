package sys_result

import (
	"net/http"

	"github.com/guojia99/cubing-pro/backend/api/internal/logic/sys_result"
	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RemoveNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoveNotifyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sys_result.NewRemoveNotifyLogic(r.Context(), svcCtx)
		resp, err := l.RemoveNotify(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
