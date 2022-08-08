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
	"strconv"
	"strings"
	"time"
)

//Messages for getting traffic messages for an area
func Messages(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if len(strings.Split(request.URL.Path, `/`)) != 6 {
		log.Println("Unable to get response for request")
		http.Error(w, "error Bad Request\nExpected input: /rtc/v1/messages/{startLocation}/{endDestination}", http.StatusBadRequest)
		return
	}
	StartAddress := strings.Split(request.URL.Path, `/`)[4] //Getting the address/name of the place we want to look for chargers
	EndAddress := strings.Split(request.URL.Path, `/`)[5]   //Getting the address/name of the place we want to look for chargers

	bodyBox, err := getBBox(StartAddress, EndAddress) //Get BBox object for traffic messages
	if err != nil {
		log.Println("Unable to get response for getBBox method provided: " + StartAddress + " and " + EndAddress + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var bbox structs.BboxStruct
	if err = json.Unmarshal(bodyBox, &bbox); err != nil { //Unmarshalling bbox response into bbox struct
		log.Println("Unable to unmarshall response into bbox struct" + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var box string
	for j := 0; j < len(bbox.Bbox); j++ { //Reformatting bbox
		coordinate := strconv.FormatFloat(bbox.Bbox[j], 'f', 6, 64)
		box += coordinate + ","
	}
	box = strings.TrimRight(box, ",")

	//Gets traffic messages in bbox area
	response, err := http.Get("https://api.tomtom.com/traffic/services/5/incidentDetails?bbox=" + url.QueryEscape(box) +
		"&fields=%7Bincidents%7Btype%2Cgeometry%7Btype%2Ccoordinates%7D%2Cproperties%7Bid%2CiconCategory%2CmagnitudeOfDelay%2Cevents%7Bdescription%2Ccode%7D%2CstartTime%2Cend" +
		"Time%2Cfrom%2Cto%2Clength%2Cdelay%2CroadNumbers%2Caci%7BprobabilityOfOccurrence%2CnumberOfReports%2ClastReportTime%7D%7D%7D%7D&key=" + utils.TomtomKey)
	err = utils.TomTomErrorHandling(response.StatusCode)
	if err != nil {
		log.Println("Unable to get response for bbox: " + "\n" + err.Error())
		http.Error(w, err.Error(), response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body) //Reads response
	if err != nil {
		log.Println("Unable to read response " + string(body) + "\n" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var messages structs.Incidents
	if err = json.Unmarshal(body, &messages); err != nil { //Unmarshalls traffic incidents into incidents struct
		log.Println("Unable to unmarshall response for body: " + string(body) + "\n" + err.Error())
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	var all []structs.OutIncident //Formats incidents
	time := time.Now().Add(-60 * time.Minute)

	for i := 0; i < len(messages.Incidents); i++ {
		if messages.Incidents[i].Properties.EndTime.Before(time) {
			startTime := messages.Incidents[i].Properties.StartTime
			endTime := messages.Incidents[i].Properties.EndTime
			FromAddress := messages.Incidents[i].Properties.From
			toAddress := messages.Incidents[i].Properties.To
			Event := messages.Incidents[i].Properties.Events[0].Description

			incidents := structs.OutIncident{From: FromAddress, To: toAddress, Start: startTime, End: endTime, Event: Event}
			all = append(all, incidents) //appends all information into incidents struct
		}
	}

	output, err := json.Marshal(all) //Marshalling the array to JSON
	if err != nil {
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		log.Println("Unable to marshall all incidents" + "\n" + err.Error())
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%v", string(output)) //Outputs the chargers

}

//getBBox function for creating a BBox object used for getting traffic messages
func getBBox(StartAddress string, endAddress string) ([]byte, error) {
	//Defines request to get latitude for startlocation
	startLat, startLong, err := database.LocationPresent(url.QueryEscape(StartAddress))
	if err != nil {
		log.Println("Unable to get response" + "\n" + err.Error())
		return nil, err
	}

	//Defines request to get latitude for endlocation
	EndLat, endLong, err := database.LocationPresent(url.QueryEscape(endAddress))
	if err != nil {
		log.Println("Unable to get response" + "\n" + err.Error())
		return nil, err
	}

	//Defines request to get BBox using startlat and endlat
	resp, err := http.Get("https://api.openrouteservice.org/v2/directions/driving-car?api_key=" + utils.OpenRouteServiceKey + "&start=" + startLong + "," + startLat + "&end=" + endLong + "," + EndLat)
	if err != nil {
		log.Println("Unable to get response" + "\n" + err.Error())
		return nil, utils.OpenRouteError(resp.StatusCode)
	}

	//Reads body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response" + "\n" + err.Error())
		return nil, err
	}

	return body, err
}
