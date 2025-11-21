package staticx

type Result struct {
	CompetitionID         string  `gorm:"column:competitionId; type:varchar(32); not null; default:''" json:"competitionId"`
	EventID               string  `gorm:"column:eventId; type:varchar(6); not null; default:''" json:"eventId"`
	RoundTypeID           string  `gorm:"column:roundTypeId; type:varchar(1); not null; default:''" json:"roundTypeId"`
	Pos                   int16   `gorm:"column:pos; type:smallint; not null; default:0" json:"pos"`
	Best                  int     `gorm:"column:best; type:int; not null; default:0" json:"best"`
	Average               int     `gorm:"column:average; type:int; not null; default:0" json:"average"`
	PersonName            *string `gorm:"column:personName; type:varchar(80)" json:"personName,omitempty"`
	PersonID              string  `gorm:"column:personId; type:varchar(10); not null; default:''" json:"personId"`
	PersonCountryID       *string `gorm:"column:personCountryId; type:varchar(50)" json:"personCountryId,omitempty"`
	FormatID              string  `gorm:"column:formatId; type:varchar(1); not null; default:''" json:"formatId"`
	Value1                int     `gorm:"column:value1; type:int; not null; default:0" json:"value1"`
	Value2                int     `gorm:"column:value2; type:int; not null; default:0" json:"value2"`
	Value3                int     `gorm:"column:value3; type:int; not null; default:0" json:"value3"`
	Value4                int     `gorm:"column:value4; type:int; not null; default:0" json:"value4"`
	Value5                int     `gorm:"column:value5; type:int; not null; default:0" json:"value5"`
	RegionalSingleRecord  *string `gorm:"column:regionalSingleRecord; type:varchar(3)" json:"regionalSingleRecord,omitempty"`
	RegionalAverageRecord *string `gorm:"column:regionalAverageRecord; type:varchar(3)" json:"regionalAverageRecord,omitempty"`
}

func (Result) TableName() string {
	return "Results"
}

type Competition struct {
	ID              string  `gorm:"column:id; type:varchar(32); not null; default:''" json:"id"`
	Name            string  `gorm:"column:name; type:varchar(50); not null; default:''" json:"name"`
	CityName        string  `gorm:"column:cityName; type:varchar(50); not null; default:''" json:"cityName"`
	CountryID       string  `gorm:"column:countryId; type:varchar(50); not null; default:''" json:"countryId"`
	Information     *string `gorm:"column:information; type:mediumtext" json:"information,omitempty"`
	Year            uint16  `gorm:"column:year; type:smallint unsigned; not null; default:0" json:"year"`
	Month           uint16  `gorm:"column:month; type:smallint unsigned; not null; default:0" json:"month"`
	Day             uint16  `gorm:"column:day; type:smallint unsigned; not null; default:0" json:"day"`
	EndMonth        uint16  `gorm:"column:endMonth; type:smallint unsigned; not null; default:0" json:"endMonth"`
	EndDay          uint16  `gorm:"column:endDay; type:smallint unsigned; not null; default:0" json:"endDay"`
	Cancelled       int     `gorm:"column:cancelled; type:int; not null; default:0" json:"cancelled"`
	EventSpecs      *string `gorm:"column:eventSpecs; type:longtext" json:"eventSpecs,omitempty"`
	WcaDelegate     *string `gorm:"column:wcaDelegate; type:mediumtext" json:"wcaDelegate,omitempty"`
	Organiser       *string `gorm:"column:organiser; type:mediumtext" json:"organiser,omitempty"`
	Venue           string  `gorm:"column:venue; type:varchar(240); not null; default:''" json:"venue"`
	VenueAddress    *string `gorm:"column:venueAddress; type:varchar(191)" json:"venueAddress,omitempty"`
	VenueDetails    *string `gorm:"column:venueDetails; type:varchar(191)" json:"venueDetails,omitempty"`
	ExternalWebsite *string `gorm:"column:external_website; type:varchar(200)" json:"externalWebsite,omitempty"`
	CellName        string  `gorm:"column:cellName; type:varchar(45); not null; default:''" json:"cellName"`
	Latitude        *int    `gorm:"column:latitude; type:int" json:"latitude,omitempty"`
	Longitude       *int    `gorm:"column:longitude; type:int" json:"longitude,omitempty"`
}

func (Competition) TableName() string {
	return "Competitions"
}
