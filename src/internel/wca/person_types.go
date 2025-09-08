package wca

type PersonBasic struct {
	Name           string        `json:"name"`
	Gender         string        `json:"gender"`
	Url            string        `json:"url"`
	WcaId          string        `json:"wca_id"`
	Id             string        `json:"id"`
	Dob            string        `json:"dob"`
	CountryIso2    string        `json:"country_iso2"`
	Class          string        `json:"class"`
	DelegateStatus interface{}   `json:"delegate_status"`
	Teams          []interface{} `json:"teams"`
}

type Country struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	ContinentId string `json:"continent_id"`
	Iso2        string `json:"iso2"`
}

type Avatar struct {
	Id               int    `json:"id"`
	Status           string `json:"status"`
	ThumbnailCropX   int    `json:"thumbnail_crop_x"`
	ThumbnailCropY   int    `json:"thumbnail_crop_y"`
	ThumbnailCropW   int    `json:"thumbnail_crop_w"`
	ThumbnailCropH   int    `json:"thumbnail_crop_h"`
	Url              string `json:"url"`
	ThumbUrl         string `json:"thumb_url"`
	IsDefault        bool   `json:"is_default"`
	CanEditThumbnail bool   `json:"can_edit_thumbnail"`
}

type PersonalRecordValue struct {
	Id            int    `json:"id"`
	PersonId      string `json:"person_id"`
	EventId       string `json:"event_id"`
	Best          int    `json:"best"`
	WorldRank     int    `json:"world_rank"`
	ContinentRank int    `json:"continent_rank"`
	CountryRank   int    `json:"country_rank"`
}

type PersonalRecord struct {
	Single  PersonalRecordValue `json:"single"`
	Average PersonalRecordValue `json:"average"`
}

type MedalCount struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

type RecordCount struct {
	National    int `json:"national"`
	Continental int `json:"continental"`
	World       int `json:"world"`
	Total       int `json:"total"`
}

type PersonProfile struct {
	Person           PersonBasic               `json:"person"`
	Country          Country                   `json:"-"` // 注意：原始数据中 country 在 person 内部，需提取
	Avatar           Avatar                    `json:"-"` // 同上，avatar 也在 person 内部
	CompetitionCount int                       `json:"competition_count"`
	PersonalRecords  map[string]PersonalRecord `json:"personal_records"`
	Medals           MedalCount                `json:"medals"`
	Records          RecordCount               `json:"records"`
}
