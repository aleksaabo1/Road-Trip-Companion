package endpoints

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Uptime of the program
var Uptime time.Time

// Diag shows diagnostics interface
// APIs in use:
// 		- https://developer.tomtom.com/
// 		- https://openrouteservice.org/
// 		- https://openweathermap.org/api
// 		- https://developer.mapquest.com
func Diag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tomtomStatusCode int
	var openRouteServiceStatusCode int
	var openWeatherMapStatusCode int
	var mapQuestStatusCode int

	// Sends a request to the TomTom API.
	respTomTom, err := http.Get("https://developer.tomtom.com/")
	// If any errors occur, log it and set the status code to StatusInternalServerError (500),
	// otherwise set the status code to the received status code (for instance StatusOK, 200).
	if err != nil {
		log.Printf("Something went wrong with the TomTom API, %v", err)
		tomtomStatusCode = http.StatusInternalServerError
	} else {
		tomtomStatusCode = respTomTom.StatusCode
		defer respTomTom.Body.Close()
	}

	// Sends a request to the OpenRouteService API.
	respOpenRouteService, err := http.Get("https://openrouteservice.org")
	// If any errors occur, log it and set the status code to StatusInternalServerError (500),
	// otherwise set the status code to the received status code (for instance StatusOK, 200).
	if err != nil {
		log.Printf("Something went wrong with the OpenRouteService API, %v", err)
		openRouteServiceStatusCode = http.StatusInternalServerError
	} else {
		openRouteServiceStatusCode = respOpenRouteService.StatusCode
		defer respOpenRouteService.Body.Close()
	}

	// Sends a request to the OpenWeatherMap API.
	respOpenWeatherMap, err := http.Get("https://api.openweathermap.org/")
	// If any errors occur, log it and set the status code to StatusInternalServerError (500),
	// otherwise set the status code to the received status code (for instance StatusOK, 200).
	if err != nil {
		log.Printf("Something went wrong with the OpenWeatherMap API, %v", err)
		openWeatherMapStatusCode = http.StatusInternalServerError
	} else {
		openWeatherMapStatusCode = respOpenWeatherMap.StatusCode
		defer respOpenWeatherMap.Body.Close()
	}

	// Sends a request to the MapQuest API.
	respMapQuest, err := http.Get("https://open.mapquestapi.com/")
	// If any errors occur, log it and set the status code to StatusInternalServerError (500),
	// otherwise set the status code to the received status code (for instance StatusOK, 200).
	if err != nil {
		log.Printf("Something went wrong with the MapQuest API, %v", err)
		mapQuestStatusCode = http.StatusInternalServerError
	} else {
		mapQuestStatusCode = respMapQuest.StatusCode
		defer respMapQuest.Body.Close()
	}

	fmt.Fprintf(w, `{"tomtom": "%v", "openrouteservice": "%v", "openweathermap": "%v", "mapquest": "%v", "version": "v1", "uptime": %v}`,
		tomtomStatusCode, openRouteServiceStatusCode, openWeatherMapStatusCode, mapQuestStatusCode, int(time.Since(Uptime)/time.Second))
}
