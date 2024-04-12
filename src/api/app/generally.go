package app

type (
	GenerallyStrIdReq struct {
		ID string `path:"id"`
	}
	GenerallyResp struct {
		Message string `json:"message"`
		Status  string `json:"status"`
		Code    int64  `json:"code"`
	}
)
type (
	GenerallyListReq struct {
		Page int64 `form:"page"`
		Size int64 `form:"size"`
	}
	GenerallyListResp struct {
		GenerallyResp
		Total int64 `json:"total"`
	}
)
