package staticx

type DNFEvent struct {
	EventID string `json:"eventID"`
	DNF     int    `json:"DNF"`
	Number  int    `json:"number"`
}

func (s *StaticX) DNFEvents(countryId string) map[string]*DNFEvent {
	var result = make(map[string]*DNFEvent)

	var allResult []Result
	if countryId != "" {
		s.db.Where("personCountryId = ?", countryId).Find(&allResult)
	} else {
		s.db.Find(&allResult)
	}

	for _, v := range allResult {
		if _, ok := result[v.EventID]; !ok {
			result[v.EventID] = &DNFEvent{
				EventID: v.EventID,
				DNF:     0,
				Number:  0,
			}
		}

		data := []int{v.Value1, v.Value2, v.Value3, v.Value4, v.Value5}
		for _, d := range data {
			if d == 0 {
				continue
			}
			if d == -2 {
				continue
			}
			result[v.EventID].Number += 1
			if d == -1 {
				result[v.EventID].DNF += 1
			}
		}
	}

	return result
}
