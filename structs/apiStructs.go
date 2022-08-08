package structs

import (
	"time"
)

type GeoLocation struct {
	Results []struct {
		Locations []struct {
			LatLng struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"latLng"`
		} `json:"locations"`
	} `json:"results"`
}

type Charger struct {
	Results []struct {
		Poi struct {
			Name  string `json:"name"`
			Phone string `json:"phone"`
		} `json:"poi,omitempty"`
		Address struct {
			FreeformAddress string `json:"freeformAddress"`
		} `json:"address,omitempty"`
		ChargingPark struct {
			Connectors []struct {
				ConnectorType string  `json:"connectorType"`
				RatedPowerKW  float64 `json:"ratedPowerKW"`
			} `json:"connectors"`
		} `json:"chargingPark,omitempty"`
	} `json:"results"`
}

type Petrol struct {
	Results []struct {
		Poi struct {
			Name   string `json:"name"`
			Brands []struct {
				Name string `json:"name"`
			} `json:"brands"`
		} `json:"poi,omitempty"`
		Address struct {
			FreeformAddress string `json:"freeformAddress"`
		} `json:"address,omitempty"`
	} `json:"results"`
}

type BboxStruct struct {
	Bbox []float64 `json:"bbox"`
}

type Incidents struct {
	Incidents []struct {
		Type       string `json:"type"`
		Properties struct {
			StartTime time.Time `json:"startTime"`
			EndTime   time.Time `json:"endTime"`
			From      string    `json:"from"`
			To        string    `json:"to"`
			Events    []struct {
				Description string `json:"description"`
			} `json:"events"`
		} `json:"properties"`
	} `json:"incidents"`
}

// WeatherData Used to store weather related data from the API, in the format the API returns the data
type WeatherData struct {
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Snow struct {
		OneH   float64 `json:"1h"`
		ThreeH float64 `json:"3h"`
	} `json:"snow"`
	Rain struct {
		OneH   float64 `json:"1h"`
		ThreeH float64 `json:"3h"`
	} `json:"rain"`
	Sys struct {
		Sunrise int `json:"sunrise"`
		Sunset  int `json:"sunset"`
	} `json:"sys"`
}

type RouteStruct struct {
	FormatVersion string `json:"formatVersion"`
	Routes        []struct {
		Summary struct {
			LengthInMeters      int       `json:"lengthInMeters"`
			TravelTimeInSeconds int       `json:"travelTimeInSeconds"`
			ArrivalTime         time.Time `json:"arrivalTime"`
		} `json:"summary"`
		Legs []struct {
			Summary struct {
				LengthInMeters int       `json:"lengthInMeters"`
				ArrivalTime    time.Time `json:"arrivalTime"`
			} `json:"summary"`
		} `json:"legs"`
		Guidance struct {
			Instructions []struct {
				Street       string   `json:"street,omitempty"`
				Maneuver     string   `json:"maneuver"`
				JunctionType string   `json:"junctionType,omitempty"`
				RoadNumbers  []string `json:"roadNumbers,omitempty"`
			} `json:"instructions"`
		} `json:"guidance"`
	} `json:"routes"`
}

type PointsOfInterest struct {
	Summary struct {
		Query        string `json:"query"`
		Querytype    string `json:"queryType"`
		Querytime    int    `json:"queryTime"`
		Numresults   int    `json:"numResults"`
		Offset       int    `json:"offset"`
		Totalresults int    `json:"totalResults"`
		Fuzzylevel   int    `json:"fuzzyLevel"`
		Geobias      struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"geoBias"`
	} `json:"summary"`
	Results []struct {
		Type  string  `json:"type"`
		ID    string  `json:"id"`
		Score float64 `json:"score"`
		Dist  float64 `json:"dist"`
		Info  string  `json:"info"`
		Poi   struct {
			Name        string `json:"name"`
			Phone       string `json:"phone"`
			Categoryset []struct {
				ID int `json:"id"`
			} `json:"categorySet"`
			Categories      []string `json:"categories"`
			Classifications []struct {
				Code  string `json:"code"`
				Names []struct {
					Namelocale string `json:"nameLocale"`
					Name       string `json:"name"`
				} `json:"names"`
			} `json:"classifications"`
		} `json:"poi,omitempty"`
		Address struct {
			Streetnumber                string `json:"streetNumber"`
			Streetname                  string `json:"streetName"`
			Municipalitysubdivision     string `json:"municipalitySubdivision"`
			Municipality                string `json:"municipality"`
			Countrysecondarysubdivision string `json:"countrySecondarySubdivision"`
			Countrysubdivision          string `json:"countrySubdivision"`
			Countrysubdivisionname      string `json:"countrySubdivisionName"`
			Postalcode                  string `json:"postalCode"`
			Extendedpostalcode          string `json:"extendedPostalCode"`
			Countrycode                 string `json:"countryCode"`
			Country                     string `json:"country"`
			Countrycodeiso3             string `json:"countryCodeISO3"`
			Freeformaddress             string `json:"freeformAddress"`
			Localname                   string `json:"localName"`
		} `json:"address"`
		Position struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"position"`
		Viewport struct {
			Topleftpoint struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"topLeftPoint"`
			Btmrightpoint struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"btmRightPoint"`
		} `json:"viewport"`
		Entrypoints []struct {
			Type     string `json:"type"`
			Position struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"position"`
		} `json:"entryPoints"`
	} `json:"results"`
}
