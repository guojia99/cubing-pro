package public_player

import (
	"net/http"

	"github.com/guojia99/cubing-pro/backend/api/internal/logic/public_player"
	"github.com/guojia99/cubing-pro/backend/api/internal/svc"
	"github.com/guojia99/cubing-pro/backend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PlayerResultHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PlayerResultReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := public_player.NewPlayerResultLogic(r.Context(), svcCtx)
		resp, err := l.PlayerResult(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
