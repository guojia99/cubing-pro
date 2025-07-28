package models

type (
	Results struct {
		EventId    string `json:"eventId"`
		Best       int    `json:"best"`
		BestStr    string `json:"bestStr"`
		Average    int    `json:"average"`
		AverageStr string `json:"averageStr"`
		PersonName string `json:"personName"`
		PersonId   string `json:"personId"`
	}

	PersonBestResults struct {
		PersonName string             `json:"PersonName"`
		Best       map[string]Results `json:"Best"`
		Avg        map[string]Results `json:"Avg"`
	}

	WCAResults struct {
		Id            int    `json:"id"`
		Best          int    `json:"best"`
		Average       int    `json:"average"`
		Name          string `json:"name"`
		CompetitionId string `json:"competition_id"`
		EventId       string `json:"event_id"`
		WcaId         string `json:"wca_id"`
		Attempts      []int  `json:"attempts"`
		BestIndex     int    `json:"best_index"`
		WorstIndex    int    `json:"worst_index"`
	}
)
