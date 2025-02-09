package setlistfm

type Setlist struct {
	ID          string `json:"id"`
	VersionID   string `json:"versionId"`
	EventDate   string `json:"eventDate"`
	LastUpdated string `json:"lastUpdated"`
	Artist      Artist `json:"artist"`
	Venue       Venue  `json:"venue"`
	Tour        Tour   `json:"tour"`
	Sets        Sets   `json:"sets"`
	Info        string `json:"info"`
	URL         string `json:"url"`
}

type Artist struct {
	Mbid           string `json:"mbid"`
	Name           string `json:"name"`
	SortName       string `json:"sortName"`
	Disambiguation string `json:"disambiguation"`
	URL            string `json:"url"`
}

type Sets struct {
	Set []Set `json:"set"`
}

type Set struct {
	Name string `json:"name"`
	Song []Song `json:"song"`
}

type Song struct {
	Name string  `json:"name"`
	Info *string `json:"info,omitempty"`
	Tape *bool   `json:"tape,omitempty"`
}

type Tour struct {
	Name string `json:"name"`
}

type Venue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	City City   `json:"city"`
	URL  string `json:"url"`
}

type City struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	State     string  `json:"state"`
	StateCode string  `json:"stateCode"`
	Coords    Coords  `json:"coords"`
	Country   Country `json:"country"`
}

type Coords struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
