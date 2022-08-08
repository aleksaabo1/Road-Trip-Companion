package utils

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// JsonUnmarshalErrorHandling Universal json unmarshalling error handler
func JsonUnmarshalErrorHandling(err error) error {
	errorString := "Unable to continue your request\n" +
		"Internal error: " + err.Error()
	return errors.New(errorString)
}

// JsonMarshalErrorHandling Universal json marshalling error handler
func JsonMarshalErrorHandling(err error) error {
	errorString := "Unable to continue your request\n" +
		"Internal error: " + err.Error()
	return errors.New(errorString)
}

// TomTomErrorHandling Universal TomTom error handler
func TomTomErrorHandling(status int) error {
	if status == http.StatusBadRequest {
		return errors.New("Error, Bad Request: " + strconv.Itoa(status) + "\nNo valid location requested.")
	} else if status == http.StatusForbidden {
		return errors.New("Error, Status Forbidden: " + strconv.Itoa(status) + "\nThe service is no longer provided.")
	} else if status == http.StatusInternalServerError {
		return errors.New("Error, Internal Server Error: " + strconv.Itoa(status) + "\nThe service is for the moment down. Please try again later.")
	} else if status == http.StatusOK {
		return nil
	}
	return errors.New("Error: " + strconv.Itoa(status) + "\n An unexpected error has occurred")
}

// OpenRouteError Universal OpenRoute error handler
func OpenRouteError(status int) error {
	if status == http.StatusBadRequest {
		return errors.New("Error, Bad Request: " + strconv.Itoa(status) + "\nNo valid location requested.")
	} else if status == http.StatusForbidden {
		return errors.New("Error, Status Forbidden: " + strconv.Itoa(status) + "\nThe service is no longer provided.")
	} else if status == http.StatusInternalServerError {
		return errors.New("Error, Internal Server Error: " + strconv.Itoa(status) + "\nThe service is for the moment down. Please try again later.")
	} else if status == http.StatusOK {
		return nil
	} else if status == http.StatusServiceUnavailable {
		return errors.New("Error, Internal Server Error: " + strconv.Itoa(status) + "\nThe service is for the moment down due to overload or maintenance. Please try again later.")
	}
	return errors.New("Error: " + strconv.Itoa(status) + "\n An unexpected error has occurred")
}

//Function to get all the filters from a url Query
func GetOptionalFilter(url *url.URL) (map[string]string, error) {
	var optionals = map[string]string{}
	optional := strings.Split(url.RawQuery, "?") //Splits the url by '?'
	if len(optional) != 0 && optional[0] != "" { //Checking if the user has passed a filter
		for i := 0; i <= len(optional)-1; i++ {
			nameOfFilter := strings.Split(optional[i], "=") //Separating the 'key' and 'value'
			if len(nameOfFilter) == 2 {
				valueName := nameOfFilter[1]   //Defining value
				mapName := nameOfFilter[0]     //Defining key
				optionals[mapName] = valueName //Adding in the map
			} else {
				return optionals, errors.New("Invalid format on filter\nMissing '=' in statement")
			}
		}
		return optionals, nil
	}
	return nil, nil
}
