package structs

import (
	"time"
)

type OutputCharge struct {
	Charger    string
	Address    string
	Phone      string
	Connectors []Connectors
}

type Connectors struct {
	ConnectorType string  `json:"connectorType"`
	RatedPowerKW  float64 `json:"ratedPowerKW"`
}

type OutputPetrol struct {
	StationName  string
	StationBrand string
	Address      string
}

type OutIncident struct {
	Start time.Time
	End   time.Time
	From  string
	To    string
	Event string
}

// OutputWeather Used to easily store and access only the wanted weather data and to add messages to the data
type OutputWeather struct {
	Main       MainStruct
	Rain1h     float64
	Snow1h     float64
	Temp       TempStruct
	FeelsLike  FeelsLikeStruct
	TempMin    TempMinStruct
	TempMax    TempMaxStruct
	Humidity   HumidityStruct
	Visibility VisibilityStruct
	WindSpeed  WindSpeedStruct
	WindDeg    WindDegStruct
	Sunrise    SunriseStruct
	Sunset     SunsetStruct
}

// MainStruct Used to add a message regarding the Main weather condition,
// which is bound to that condition
type MainStruct struct {
	Main    string
	Message string
}

// TempStruct Used to add a message regarding the temperature,
// which is bound to that temperature and changes with the temperature
type TempStruct struct {
	Temp    float64
	Message string
}

// FeelsLikeStruct Used to add a message regarding the feels-like temperature,
// which floats and changes with the feels-like temperature
type FeelsLikeStruct struct {
	FeelsLike float64
	Message   string
}

// TempMinStruct Used to add a message regarding the minimum temperature,
// and floats and changes with the minimum temperature
type TempMinStruct struct {
	TempMin float64
	Message string
}

// TempMaxStruct Used to add a message regarding the maximum temperature,
// and floats and changes with the maximum temperature
type TempMaxStruct struct {
	TempMax float64
	Message string
}

// HumidityStruct Used to add message to the humidity value, which changes if the humidity changes
type HumidityStruct struct {
	Humidity int
	Message  string
}

// VisibilityStruct Used to add message to the visibility value and changes if the visibility changes.
type VisibilityStruct struct {
	Visibility int
	Message    string
}

// WindSpeedStruct Used to add message to the wind speed, the message will change if the wind speed
// increase or decrease.
type WindSpeedStruct struct {
	WindSpeed float64
	Message   string
}

// WindDegStruct Used to add a "floating" message to the wind deg.
type WindDegStruct struct {
	WindDeg int
	Message string
}

// SunriseStruct Used to add message to the sunrise time, adds human readable sunrise value as message.
type SunriseStruct struct {
	Sunrise int
	Message string
}

// SunsetStruct Used to add message to the sunset time, adds human readable sunset value as message.
type SunsetStruct struct {
	Sunset  int
	Message string
}

type Route struct {
	Street       string
	Maneuver     string
	RoadNumber   string
	JunctionType string
}

type RoadInformation struct {
	EstimatedArrival string
	LengthKM         int
	Route            []Route
}

type Webhook struct {
	Id                  string
	Url                 string
	DepartureLocation   string
	ArrivalDestination  string
	Weather             string
	ArrivalTime         string
	EstimatedTravelTime int
}

type NotificationInput struct {
	URL string
}

type NotificationResponse struct {
	Text string
}

type OutputPoi struct {
	Name        string
	PhoneNumber string
	Address     string
}

type LocationLonLat struct {
	Longitude string
	Latitude  string
}

type JsonMessage struct {
	Text       string        `json:"text"`
	Attachment []Attachments `json:"attachments"`
}

type Attachments struct {
	Color      string `json:"color"`
	AuthorName string `json:"author_name"`
	Title      string `json:"title"`
	TitleLink  string `json:"title_link"`
	Text       string `json:"text"`
	Footer     string `json:"footer"`
}
