package endpoints

import (
	"cloudproject/database"
	"cloudproject/structs"
	"cloudproject/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// PointOfInterest Displays all the points of interest from a location, within 5 km radius
func PointOfInterest(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//Getting the address/name of the place we want to look for points of interest
	address := strings.Split(request.URL.Path, `/`)[4]

	//Gets the interest the user are interested in
	poiPath := strings.Split(request.URL.Path, `/`)[5]

	//Receives the latitude and longitude of the place passed in to the url
	latitude, longitude, err := database.LocationPresent(url.QueryEscape(address))
	if err != nil {
		log.Println("There was an error retrieving the GeoCode for location: " + address)
		http.Error(w, "ERROR, The searched place does not exist", http.StatusBadRequest)
		return
	}

	// Sends a GET request to the API and stores the response
	response, err := http.Get("https://api.tomtom.com/search/2/poiSearch/" + poiPath + ".json?lat=" + latitude + "&lon=" + longitude + "&radius=5000&key=gcP26xVobGHjX2VVWGTskjelxX81WA1G")

	// Reads the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("An error occurred while reading the response body.\n" + err.Error())
		http.Error(w, "ERROR, The searched point of interest does not exist", http.StatusBadRequest)
		return
	}

	var poi structs.PointsOfInterest

	if err = json.Unmarshal(body, &poi); err != nil {
		log.Println("An error occurred during unmarshal.\n" + err.Error())
		unmarshalErr := utils.JsonUnmarshalErrorHandling(err)
		http.Error(w, unmarshalErr.Error(), http.StatusInternalServerError)
		return
	}

	var total []structs.OutputPoi

	// For each point of interest
	for i := 0; i < len(poi.Results); i++ {
		// Store the data from this point of interest into the different struct fields
		poiName := poi.Results[i].Poi.Name
		poiPhoneNumber := poi.Results[i].Poi.Phone
		poiAddress := poi.Results[i].Address.Freeformaddress

		jsonStruct := structs.OutputPoi{Name: poiName, PhoneNumber: poiPhoneNumber, Address: poiAddress} //Creating a JSON object
		total = append(total, jsonStruct)                                                                //Appending the json object to an array
	}

	output, err := json.Marshal(total) //Marshaling the array to JSON
	if err != nil {
		log.Println("An error occurred during unmarshal.\n" + err.Error())
		marshalErr := utils.JsonMarshalErrorHandling(err)
		http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
		return
	}

	// Display the output to the user
	_, err = fmt.Fprintf(w, "%v", string(output))
	if err != nil {
		log.Println("There has been an error displaying the data to the user.")
		http.Error(w, "There has been an error when displaying the data.", http.StatusInternalServerError)
		return
	}

}
