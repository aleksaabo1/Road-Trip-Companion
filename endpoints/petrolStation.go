package endpoints

import (
	"cloudproject/database"
	"cloudproject/structs"
	"cloudproject/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//PetrolStation Function that will display all the electric-vehicle charging stations from a location, within 1km
func PetrolStation(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	address := strings.Split(request.URL.Path, `/`)[4] //Getting the address/name of the place we want to look for chargers
	if address == "" {
		http.Error(w, "Please insert a Location", http.StatusBadRequest)
		return
	}

	latitude, longitude, err := database.LocationPresent(url.QueryEscape(address)) //Receives the latitude and longitude of the place passed in the url
	if err != nil {
		log.Println("Unable to retrieve latitude and longitude for location: " + address + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter, err := utils.GetOptionalFilter(request.URL) //Getting the optional filters
	if err != nil {
		log.Println("Unable to retrieve filters from url: " + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var response *http.Response //Defines a response object

	if len(filter) != 0 {
		radius, err := checkFilter(filter) //Getting filters
		if err != nil {
			log.Println("Unable to check filters for filter " + "\n" + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//http.get with radius using the latitude and longitude
		response, err = http.Get("https://api.tomtom.com/search/2/nearbySearch/.json?lat=" + latitude + "&lon=" + longitude + radius + "&categorySet=7311&fuelSet=&key=" + utils.TomtomKey)
	} else {
		//http.get without user specified radius
		response, err = http.Get("https://api.tomtom.com/search/2/nearbySearch/.json?lat=" + latitude + "&lon=" + longitude + "&radius=5000&categorySet=7311&key=" + utils.TomtomKey)
	}

	body, err := ioutil.ReadAll(response.Body) //Reading body
	if err != nil {
		log.Println("Unable to read body for body:" + string(body) + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}

	var petrol structs.Petrol
	if err = json.Unmarshal(body, &petrol); err != nil { //Unmarshalling the body to json form
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		log.Println("Unable to unmarshall body into petrol for body: " + string(body) + "\n" + err.Error())
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	var total []structs.OutputPetrol
	for i := 0; i < len(petrol.Results); i++ { //For each of the stations

		stationName := petrol.Results[i].Poi.Name //Variable storing the station name
		var stationBrand string
		if len(petrol.Results[i].Poi.Brands) != 0 {
			stationBrand = petrol.Results[i].Poi.Brands[0].Name //Variable storing the brand name
		}
		address := petrol.Results[i].Address.FreeformAddress //Getting the address to the station

		jsonStruct := structs.OutputPetrol{StationName: stationName, StationBrand: stationBrand, Address: address} //Creating a JSON object
		total = append(total, jsonStruct)                                                                          //Appending the json object to an array
	}

	output, err := json.Marshal(total) //Marshalling the array to JSON
	if err != nil {
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		log.Println("Unable marshall object, output: " + string(output) + "\n" + err.Error())
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%v", string(output)) //Outputs the chargers

}

//checkFilter Function to check if the filter is valid and has the proper input
func checkFilter(filter map[string]string) (string, error) {
	_, foundRadius := filter["radius"]
	//If statement to check if the user passed in a correct filter, and with a value
	if !(foundRadius) {
		return "", errors.New("error, Bad Request\nNone of the filters is accepted\nAccepted filters: radius, charge, power")
	} else if len(filter["radius"]) == 0 { //Checking if the user has passed in a valid filter
		return "", errors.New("error, Bad Request\nField cannot be empty")
	}
	radius := ""
	if len(filter["radius"]) != 0 {
		//Checks if the user has passed in an int, and not a string
		if _, err := strconv.Atoi(filter["radius"]); err != nil {
			return "", errors.New("Value of radius must be a number\nTry again")
		} else {
			radius = "&radius=" + filter["radius"]
		}
	}
	return radius, nil
}
