package acknowledgments

const thanksKey = "__thanks_with_ui_acknowledgments__"

type Thank struct {
	WcaID    string  `json:"wcaID"`
	Nickname string  `json:"nickname"`
	Amount   float64 `json:"amount"`
	Avatar   string  `json:"avatar"`
	Other    string  `json:"other"`
}

type Thanks []Thank
