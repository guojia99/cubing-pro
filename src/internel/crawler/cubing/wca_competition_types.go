package cubing

import (
	"time"
)

type WCAInfoPerson struct {
	Name         string      `json:"name"`
	WcaUserId    int         `json:"wcaUserId"`
	WcaId        *string     `json:"wcaId"`
	RegistrantId interface{} `json:"registrantId"`
	CountryIso2  string      `json:"countryIso2"`
	Gender       string      `json:"gender"`
	Registration interface{} `json:"registration"`
	Avatar       *struct {
		Url      string `json:"url"`
		ThumbUrl string `json:"thumbUrl"`
	} `json:"avatar"`
	Roles         []string      `json:"roles"`
	Assignments   []interface{} `json:"assignments"`
	PersonalBests []struct {
		EventId            string `json:"eventId"`
		Best               int    `json:"best"`
		WorldRanking       int    `json:"worldRanking"`
		ContinentalRanking int    `json:"continentalRanking"`
		NationalRanking    int    `json:"nationalRanking"`
		Type               string `json:"type"`
	} `json:"personalBests"`
	Extensions []interface{} `json:"extensions"`
}

type (
	WCAInfoEvent struct {
		Id            string              `json:"id"`
		Rounds        []WCAInfoEventRound `json:"rounds"`
		Extensions    []interface{}       `json:"extensions"`
		Qualification interface{}         `json:"qualification"`
	}

	WCAInfoEventRoundTimeLimit struct {
		Centiseconds       int           `json:"centiseconds"`
		CumulativeRoundIds []interface{} `json:"cumulativeRoundIds"`
	}

	WCAInfoEventRoundCutoff struct {
		NumberOfAttempts int `json:"numberOfAttempts"`
		AttemptResult    int `json:"attemptResult"`
	}

	WCAInfoEventRoundAdvancementCondition struct {
		Type  string `json:"type"`
		Level int    `json:"level"`
	}

	WCAInfoEventRound struct {
		Id                   string                                 `json:"id"`
		Format               string                                 `json:"format"`
		TimeLimit            WCAInfoEventRoundTimeLimit             `json:"timeLimit"`
		Cutoff               *WCAInfoEventRoundCutoff               `json:"cutoff"`
		AdvancementCondition *WCAInfoEventRoundAdvancementCondition `json:"advancementCondition"`
		ScrambleSetCount     int                                    `json:"scrambleSetCount"`
	}
)

type WCAInfoSchedule struct {
	StartDate    string `json:"startDate"`
	NumberOfDays int    `json:"numberOfDays"`
	Venues       []struct {
		Id                    int    `json:"id"`
		Name                  string `json:"name"`
		LatitudeMicrodegrees  int    `json:"latitudeMicrodegrees"`
		LongitudeMicrodegrees int    `json:"longitudeMicrodegrees"`
		CountryIso2           string `json:"countryIso2"`
		Timezone              string `json:"timezone"`
		Rooms                 []struct {
			Id         int    `json:"id"`
			Name       string `json:"name"`
			Color      string `json:"color"`
			Activities []struct {
				Id              int           `json:"id"`
				Name            string        `json:"name"`
				ActivityCode    string        `json:"activityCode"`
				StartTime       time.Time     `json:"startTime"`
				EndTime         time.Time     `json:"endTime"`
				ChildActivities []interface{} `json:"childActivities"`
				Extensions      []interface{} `json:"extensions"`
			} `json:"activities"`
			Extensions []interface{} `json:"extensions"`
		} `json:"rooms"`
		Extensions []interface{} `json:"extensions"`
	} `json:"venues"`
}

type WCAInfoRegistrationInfo struct {
	OpenTime              time.Time `json:"openTime"`
	CloseTime             time.Time `json:"closeTime"`
	BaseEntryFee          int       `json:"baseEntryFee"`
	CurrencyCode          string    `json:"currencyCode"`
	OnTheSpotRegistration bool      `json:"onTheSpotRegistration"`
	UseWcaRegistration    bool      `json:"useWcaRegistration"`
}

type WCAInfo struct {
	FormatVersion    string                  `json:"formatVersion"`
	Id               string                  `json:"id"`
	Name             string                  `json:"name"`
	ShortName        string                  `json:"shortName"`
	Persons          []WCAInfoPerson         `json:"persons"`
	Events           []WCAInfoEvent          `json:"events"`
	Schedule         WCAInfoSchedule         `json:"schedule"`
	CompetitorLimit  int                     `json:"competitorLimit"`
	RegistrationInfo WCAInfoRegistrationInfo `json:"registrationInfo"`
}
