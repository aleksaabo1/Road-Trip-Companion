package endpoints

import (
	"cloudproject/database"
	structs2 "cloudproject/structs"
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

// EVStations Displays all the electric-vehicle charging stations from a location, within 5 km
func EVStations(w http.ResponseWriter, request *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	//Getting the address/name of the place we want to look for chargers
	address := strings.Split(request.URL.Path, `/`)[4] //Getting the address/name of the place we want to look for chargers

	if address == "" {
		log.Println("There is no address provided.")
		http.Error(w, "Please insert a Location", http.StatusBadRequest)
		return
	}

	//Receives the latitude and longitude of the place passed in to the url
	latitude, longitude, err := database.LocationPresent(url.QueryEscape(address))
	if err != nil {
		log.Println("Unable to retrieve GeoCode for location: " + address + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If there is an optional filter, retrieve it
	filter, err := utils.GetOptionalFilter(request.URL)
	if err != nil {
		log.Println("Unable to retrieve filter(s).\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response *http.Response

	// No filters provided
	if len(filter) != 0 {
		// Return the filter
		connector, power, radius, err := checkOptional(filter)
		if err != nil {
			log.Println("There was an error while retrieving filters.\n" + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Get method with filters
		response, err = http.Get("https://api.tomtom.com/search/2/nearbySearch/.json?lat=" + latitude + "&lon=" + longitude + radius + "&connectorSet=" + connector + power + "&categorySet=7309&key=" + utils.TomtomKey)
	} else {
		// Get method without filters
		response, err = http.Get("https://api.tomtom.com/search/2/nearbySearch/.json?lat=" + latitude + "&lon=" + longitude + "&radius=5000&categorySet=7309&key=" + utils.TomtomKey)
	}

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("There was an error while reading the response body.\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}

	var charge structs2.Charger

	// Unmarshalling the body
	if err = json.Unmarshal(body, &charge); err != nil {
		log.Println("There was an error during unmarshalling.\n" + err.Error())
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	var total []structs2.OutputCharge
	for i := 0; i < len(charge.Results); i++ {
		addressCharge := charge.Results[i].Address.FreeformAddress //Address where the ev station is located
		chargeName := charge.Results[i].Poi.Name                   //Name of the charger
		phone := charge.Results[i].Poi.Phone                       //Phone number to charger maintainer

		var connector string
		var power float64
		var connectorStruct []structs2.Connectors
		if len(charge.Results[i].ChargingPark.Connectors) != 0 {

			// Nested for-loop
			// Gets the connector types and minimum power for each charging station
			for j := 0; j < len(charge.Results[i].ChargingPark.Connectors); j++ {
				connector = charge.Results[i].ChargingPark.Connectors[j].ConnectorType           //The connector types
				power = charge.Results[i].ChargingPark.Connectors[j].RatedPowerKW                //The minimum power
				connectors := structs2.Connectors{ConnectorType: connector, RatedPowerKW: power} //Creating a struct
				connectorStruct = append(connectorStruct, connectors)                            //Adding a struct to an array of Connectors struct
			}
		}

		jsonStruct := structs2.OutputCharge{Charger: chargeName, Address: addressCharge, Phone: phone, Connectors: connectorStruct} //Creating a JSON object
		total = append(total, jsonStruct)                                                                                           //Appending the json object to an array
	}

	//Checking if the struct is empty
	if total == nil {
		log.Println("The json struct is empty.")
		http.Error(w, "No electric charges in this area", http.StatusNoContent)
		return
	}

	// Marshalling the array to JSON
	output, err := json.Marshal(total)
	if err != nil {
		log.Println("There was an error while marshalling the data.\n" + err.Error())
		jsonError := utils.JsonMarshalErrorHandling(err)
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
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

// checkOptional Checks if the filter is valid and has the proper input
func checkOptional(filter map[string]string) (string, string, string, error) {
	_, foundCharge := filter["connector"]
	_, foundRadius := filter["radius"]
	_, foundPower := filter["power"]

	// If statement to check if the user passed in a correct filter, and with a value
	if !(foundCharge || foundPower || foundRadius) {
		return "", "", "", errors.New("error, Bad Request\nNone of the filters is accepted\nAccepted filters: radius, charge, power")
	} else if len(filter["connector"]) == 0 && len(filter["radius"]) == 0 && len(filter["power"]) == 0 {
		return "", "", "", errors.New("error, Bad Request\nField cannot be empty")
	}

	connector := ""
	power := ""
	radius := "&radius=5000" //Hardcoded value, to satisfy the url, if the user has not passed in the radius
	if len(filter["connector"]) != 0 {
		chargingOutlet := outletSearch(filter["connector"]) //Checks if the user has passed in a correct connector outlet
		if chargingOutlet != "" {
			connector = "&connectorSet=" + filter["connector"] //Format the filter to support api url
		} else {
			return "", "", "", errors.New("Connector Not supported\nThe connector is not supported in our system")
		}
	}
	if len(filter["radius"]) != 0 {
		if _, err := strconv.Atoi(filter["radius"]); err != nil { //Checks if the user has passed in an int, and not a string
			return "", "", "", errors.New("Value of radius must be a number\nTry again")
		} else {
			radius = "&radius=" + filter["radius"] //Format the filter to support api url
		}
	}
	if len(filter["power"]) != 0 {
		if _, err := strconv.Atoi(filter["power"]); err != nil { //Checks if the user has passed in an int, and not a string
			return "", "", "", errors.New("Value of power must be a number\nTry again")
		} else {
			power = "&minPowerKW=" + filter["power"] //Format the filter to support api url
		}
	}
	return connector, power, radius, nil
}

// outletsMap Map with alternative searches, that will map to the supported name to the API
var outletsMap = map[string]string{
	"standard":    "StandardHouseholdCountrySpecific",
	"type1":       "IEC62196Type1",
	"IEC62196-2":  "IEC62196Type1",
	"type1ccs":    "IEC62196Type1CCS",
	"IEC62196-3":  "IEC62196Type1CCS",
	"type2":       "IEC62196Type2Outlet",
	"typeccs":     "IEC62196Type2CCS",
	"type3":       "IEC62196Type3",
	"chademo":     "Chademo",
	"gbt202342":   "GBT20234Part2",
	"gbt202343":   "GBT20234Part3",
	"IEC60309AC3": "IEC60309AC3PhaseRed",
	"PhaseRed":    "IEC60309AC3PhaseRed",
	"IEC60309AC1": "IEC60309AC1PhaseBlue",
	"PhaseBlue":   "IEC60309AC1PhaseBlue",
	"IEC60309":    "IEC60309DCWhite",
	"DCWhite":     "IEC60309DCWhite",
}

// outletArray Array with the supported chargers to the TomTom API
var outletArray = []string{
	"StandardHouseholdCountrySpecific", "IEC62196Type1", "IEC62196Type1CCS", "IEC62196Type2CableAttached",
	"IEC62196Type2Outlet", "IEC62196Type2CCS", "IEC62196Type3", "Chademo", "GBT20234Part2", "GBT20234Part3",
	"IEC60309AC3PhaseRed", "IEC60309AC1PhaseBlue", "IEC60309DCWhite", "Tesla",
}

// outletSearch Checks if the user has passed in a supported connector type
func outletSearch(searchedOutlet string) string {
	for i := 0; i < len(outletArray); i++ {
		// If the user passed in a connector with the correct syntax from the API
		if strings.ToLower(searchedOutlet) == strings.ToLower(outletArray[i]) {
			return outletArray[i]
		}
	}
	//If the user has passed in a value from the coded map
	if outlet, found := outletsMap[searchedOutlet]; found {
		return outlet
	}

	return ""

}
