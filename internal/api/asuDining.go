package api

type StationName string

var (
	HomeStationName        StationName = "Home Zone 1"
	TrueBalanceStationName StationName = "True Balance"
	SoupStationName        StationName = "Soup Station"
)

type Response struct {
	Menu Menu `json:"Menu"`
}

type Menu struct {
	MenuPeriods  []MenuPeriod  `json:"MenuPeriods"`
	MenuProducts []MenuProduct `json:"MenuProducts"`
	MenuStations []MenuStation `json:"MenuStations"`
}

type MenuPeriod struct {
	PeriodID string `json:"PeriodId"`
	Name     string `json:"Name"`
}

type MenuProduct struct {
	PeriodID  string  `json:"PeriodId"`
	StationID string  `json:"StationId"`
	Product   Product `json:"Product"`
}

type Product struct {
	MarketingName    string `json:"MarketingName"`
	ShortDescription string `json:"ShortDescription"`
	Calories         string `json:"Calories"`
}

type MenuStation struct {
	PeriodID  string `json:"PeriodId"`
	StationID string `json:"StationId"`
	Name      string `json:"Name"`
}

type Dish struct {
	Name     string
	Calories string
}
