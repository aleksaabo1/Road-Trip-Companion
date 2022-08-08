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

//Route function will respond with a route from the specified location to a destination
func Route(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(strings.Split(request.URL.Path, `/`)) != 6 { //Splits url in order to get startLocation and endLocation
		http.Error(w, "error Bad Request\nExpected input: /route/{startLocation}/{endDestination}", http.StatusBadRequest)
		return
	}

	StartAddress := strings.Split(request.URL.Path, `/`)[4] //Getting the address/name of the place we want to look for chargers
	EndAddress := strings.Split(request.URL.Path, `/`)[5]   //Getting the address/name of the place we want to look for chargers

	startLat, startLong, err := database.LocationPresent(url.QueryEscape(StartAddress)) //Gets coordinates of start address
	if err != nil {
		log.Println("Unable to get request for address: " + StartAddress + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	EndLat, endLong, err := database.LocationPresent(url.QueryEscape(EndAddress)) //Gets coordinates for destination
	if err != nil {
		log.Println("Unable to get request for address: " + EndAddress + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	coordinates := startLat + "%2C" + startLong + "%3A" + EndLat + "%2C" + endLong

	//Gets route using coordinates of start and end location
	resp, err := http.Get("https://api.tomtom.com/routing/1/calculateRoute/" + coordinates + "/json?instructionsType=coded&traffic=false&avoid=unpavedRoads&travelMode=car&key=" + utils.TomtomKey)
	errTomTom := utils.TomTomErrorHandling(resp.StatusCode)
	if err != nil {
		log.Println("Unable to get response for coordinates: " + coordinates + "\n" + err.Error())
		http.Error(w, err.Error(), resp.StatusCode)
		return
	} else if errTomTom != nil {
		log.Println("TomTom error for coordinates: " + coordinates + "\n" + err.Error())
		http.Error(w, errTomTom.Error(), resp.StatusCode)
		return
	}

	//Reads body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read body:" + string(body) + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	//Unmarshalls response into a roads object
	var roads structs.RouteStruct
	if err = json.Unmarshal(body, &roads); err != nil {
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		log.Println("Unable to unmarshal response for body: " + string(body) + "\n" + err.Error())
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	//Defines variables and structs from roads object
	var maneuver string
	var junctionType string
	var RoadNumber string
	var Street string

	var total []structs.Route

	drivingLength := roads.Routes[0].Summary.LengthInMeters / 1000
	estimatedTime := roads.Routes[0].Summary.ArrivalTime
	estimatedTimeString := estimatedTime.Format("2006-01-02 15:04:05")

	//For each instruction get maneuver and roadnumber
	for i := 0; i < len(roads.Routes[0].Guidance.Instructions); i++ {
		maneuver = roads.Routes[0].Guidance.Instructions[i].Maneuver
		maneuver = maneuvers[maneuver]
		junctionType = roads.Routes[0].Guidance.Instructions[i].JunctionType
		if roads.Routes[0].Guidance.Instructions[i].RoadNumbers != nil {
			RoadNumber = roads.Routes[0].Guidance.Instructions[i].RoadNumbers[0]
		}

		//Puts instructions into street object
		Street = roads.Routes[0].Guidance.Instructions[i].Street

		route := structs.Route{Street: Street, RoadNumber: RoadNumber, Maneuver: maneuver, JunctionType: junctionType}
		total = append(total, route) //Appends information
	}

	information := structs.RoadInformation{EstimatedArrival: estimatedTimeString, LengthKM: drivingLength, Route: total}

	output, err := json.Marshal(information) //Marshalling the array to JSON
	if err != nil {
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		log.Println("Unable to marshall response: " + "\n" + err.Error())
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%v", string(output)) //Outputs the chargers

}

//Maneuvers that map to a more detailed description
var maneuvers = map[string]string{
	"ARRIVE":               "You have arrived.",
	"ARRIVE_LEFT":          "You have arrived. Your destination is on the left.",
	"ARRIVE_RIGHT":         "You have arrived. Your destination is on the right.",
	"DEPART":               "Leave.",
	"STRAIGHT":             "Keep straight on.",
	"KEEP_RIGHT":           "Keep right.",
	"BEAR_RIGHT":           "Bear right.",
	"TURN_RIGHT":           "Turn right.",
	"SHARP_RIGHT":          "Turn sharp right.",
	"KEEP_LEFT":            "Keep left.",
	"BEAR_LEFT":            "Bear left.",
	"TURN_LEFT":            "Turn left.",
	"SHARP_LEFT":           "Turn sharp left.",
	"MAKE_UTURN":           "Make a U-turn.",
	"ENTER_MOTORWAY":       "Take the motorway.",
	"ENTER_FREEWAY":        "Take the freeway.",
	"ENTER_HIGHWAY":        "Take the highway.",
	"TAKE_EXIT":            "Take the exit.",
	"MOTORWAY_EXIT_LEFT":   "Take the left exit.",
	"MOTORWAY_EXIT_RIGHT":  "Take the right exit.",
	"TAKE_FERRY":           "Take the ferry.",
	"ROUNDABOUT_CROSS":     "Cross the roundabout.",
	"ROUNDABOUT_RIGHT":     "At the roundabout take the exit on the right.",
	"ROUNDABOUT_LEFT":      "At the roundabout take the exit on the left.",
	"ROUNDABOUT_BACK":      "Go around the roundabout.",
	"TRY_MAKE_UTURN":       "Try to make a U-turn.",
	"FOLLOW":               "Follow.",
	"SWITCH_PARALLEL_ROAD": "Switch to the parallel road.",
	"SWITCH_MAIN_ROAD":     "Switch to the main road.",
	"ENTRANCE_RAMP":        "Take the ramp.",
	"WAYPOINT_LEFT":        "You have reached the waypoint. It is on the left.",
	"WAYPOINT_RIGHT":       "You have reached the waypoint. It is on the right.",
	"WAYPOINT_REACHED":     "You have reached the waypoint."}
