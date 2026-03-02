package types

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

type PersonResult struct {
	EventId       string `json:"eventId"`
	Best          int    `json:"best,omitempty"`
	BestStr       string `json:"bestStr,omitempty"`
	PersonName    string `json:"personName,omitempty"`
	PersonId      string `json:"personId"` // wcaId
	WorldRank     int    `json:"world_rank"`
	ContinentRank int    `json:"continent_rank"`
	CountryRank   int    `json:"country_rank"`

	Rank  int    `json:"rank"` // diy rank
	Times string `json:"times,omitempty"`
}

type PersonalRecord struct {
	Best *PersonResult `json:"single"`
	Avg  *PersonResult `json:"average"`
}

type PersonInfo struct {
	PersonName       string                    `json:"name"`
	WcaID            string                    `json:"wcaId"`
	CountryID        string                    `json:"countryId"`
	Gender           string                    `json:"gender"`
	CountryIso2      string                    `json:"country_iso2"`
	PersonalRecords  map[string]PersonalRecord `json:"personal_records"`
	CompetitionCount int                       `json:"competition_count"`
	MedalCount       MedalCount                `json:"medalCount"`
	RecordCount      RecordCount               `json:"recordCount"`
}

type (
	PersonBestRank struct {
		Best map[string]PersonResult `json:"best"`
		Avg  map[string]PersonResult `json:"average"`
	}

	PersonBestRanks struct {
		//All    PersonBestRank `json:"all"`
		WithNR PersonBestRank `json:"withNR"`
		WithCR PersonBestRank `json:"withCR"`
		WithWR PersonBestRank `json:"withWR"`
	}
)
