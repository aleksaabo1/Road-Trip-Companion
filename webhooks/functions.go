package webhooks

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"cloudproject/database"
	"cloudproject/endpoints"
	"cloudproject/structs"
	"cloudproject/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SignatureKey Initialize signature
var SignatureKey = "X-SIGNATURE"

// Secret byte array
var Secret []byte

// CalculateDeparture Calculates the time of departure based on weather conditions and traffic messages
// (The traffic messages is considered by the API it self, but has an impact on the time it takes from one
// destination to another)
func CalculateDeparture(id string) error {
	// Retrieves the webhook and its information from the database
	webhookInformation, _ := database.Client.Collection(database.Collection).Doc(id).Get(database.Ctx)

	// Defines instance of Webhook-struct
	var message structs.Webhook
	// Tries to input the data from the webhook in the database, into to the struct
	if err := webhookInformation.DataTo(&message); err != nil {
		log.Println(err.Error())
		return errors.New("internal error, could not calculate time, try again")
	}

	// Retrieves the latitude and longitude for the departure location
	startLat, startLong, err := database.LocationPresent(url.QueryEscape(message.DepartureLocation))
	if err != nil {
		log.Println("There was an error retrieving the departure locations latitude and longitude." +
			"\n" + err.Error())
		return errors.New("internal error, could not calculate time, try again")

	}

	endLat, endLong, err := database.LocationPresent(url.QueryEscape(message.ArrivalDestination))
	if err != nil {
		log.Println("There was an error retrieving the arrival destinations latitude and longitude." +
			"\n" + err.Error())
		return errors.New("internal error, could not calculate time, try again")

	}

	// Stores the latitudes and longitudes for the start- and end location to be used in the call to the API.
	// Have to use '%2C' for ',' and '%3A' for ':'
	coordinates := startLat + "%2C" + startLong + "%3A" + endLat + "%2C" + endLong

	// Sends a Get-request to the API to call for route data such as travel time (as we need in this instance)
	resp, err := http.Get("https://api.tomtom.com/routing/1/calculateRoute/" + coordinates + "/json?instructionsType=coded&traffic=false&avoid=unpavedRoads&travelMode=car&key=" + utils.TomtomKey)
	if err != nil {
		log.Println("There was an error retrieving travel data from the TomTom API, Status Code: " + strconv.Itoa(http.StatusInternalServerError) +
			"\n" + err.Error())
		err = utils.TomTomErrorHandling(resp.StatusCode)
		return errors.New("internal error, could not calculate time, try again")

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("There was an error reading the response body, Status Code: " + strconv.Itoa(http.StatusInternalServerError) +
			"\n" + err.Error())
		return errors.New("internal error, could not calculate time, try again")

	}

	var roads structs.RouteStruct
	if err = json.Unmarshal(body, &roads); err != nil {
		unmarshalErr := utils.JsonUnmarshalErrorHandling(err)
		log.Println(unmarshalErr)
		return errors.New("internal error, could not calculate time, try again")
	}

	estimatedTravelTime := roads.Routes[0].Summary.TravelTimeInSeconds

	// Estimates the travel time based on the actual travel time provided by the API, and adds the weighted time which
	// is calculated using weather conditions.
	estimatedTravelTimeMinutes := (estimatedTravelTime + endpoints.GetMessageWeight(message.Weather)) / 60

	// Updates the estimated travel time for the webhook in the database by setting the newly calculated travel time
	// as the travel time.
	_, err = database.Client.Collection(database.Collection).Doc(id).Set(database.Ctx, map[string]interface{}{
		"EstimatedTravelTime": estimatedTravelTimeMinutes,
	}, firestore.MergeAll)

	return nil
}

// CallUrl Calls the URL provided in the webhook on invocation
func CallUrl(url string, content string) {

	// Creates a POST request with the content
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(content)))
	if err != nil {
		log.Println("Error during request creation - unable to notify on URL: " + url)
		return
	}

	// Hash content using sha256
	mac := hmac.New(sha256.New, Secret)
	_, errHash := mac.Write([]byte(content))
	if errHash != nil {
		log.Println("Error during content hash.")
		_ = fmt.Errorf("%v", "Error during content hashing.")
		return
	}
	// Convert to string & add to header
	req.Header.Add(SignatureKey, hex.EncodeToString(mac.Sum(nil)))

	client := http.Client{}

	// Send the request
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error in HTTP request: " + err.Error())
		return
	}

	// Reading response
	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Something is wrong with invocation response, Status Code: " + strconv.Itoa(res.StatusCode) +
			"\n" + err.Error())
		return
	}

	log.Println("Webhook invoked. Received Status Code: " + strconv.Itoa(res.StatusCode) +
		" and body: " + string(response))
}

// SendNotification Creates POST body which is supported by Slack and controls when to invoke the webhooks
func SendNotification(notificationId string) {
	// Checks through all entries in collection "messages" for a webhook with id: notificationId
	doc, err := database.Client.Collection(database.Collection).Doc(notificationId).Get(database.Ctx)
	if err != nil {
		log.Println("Unable to find webhook with ID: " + notificationId + " in the " + database.Collection + " collection")
		_ = errors.New("The notification ID is not in our system")
		return
	}
	var firebase structs.Webhook
	var arrivalTime string
	var notificationUrl string
	var timeUntilInvocation float64

	// Tries to add the data from firebase to the Webhook-struct
	if err := doc.DataTo(&firebase); err != nil {
		log.Println("Could not add webhook data to struct. \n" + err.Error())
		return
	}

	notificationUrl = firebase.Url
	arrivalTime = firebase.ArrivalTime
	destination := firebase.ArrivalDestination

	isValid := utils.IsValidInput(arrivalTime)
	var newTime time.Time
	if isValid == true {
		timeS, _ := time.Parse(time.RFC822, firebase.ArrivalTime)
		newTime = timeS.Add(time.Duration(-firebase.EstimatedTravelTime-30) * time.Minute)
		timeUntilInvocation = time.Until(newTime).Minutes()
		if timeUntilInvocation < 0 {
			return
		}
	} else {
		log.Println("Error when parsing arrival time to RFC822-format.")
		return
	}

	// Sleeps the go-routine for a given time
	time.Sleep(time.Duration(timeUntilInvocation) * time.Minute)

	//Getting the updated weather
	doc, err = database.Client.Collection(database.Collection).Doc(notificationId).Get(database.Ctx)
	if err != nil {
		log.Println("Unable to find webhook with ID: " + notificationId + " in the " + database.Collection + " collection")
		_ = errors.New("The notification ID is not in our system")
		return
	}

	// Tries to add the data from firebase to the Webhook-struct
	if err := doc.DataTo(&firebase); err != nil {
		log.Println("Could not add webhook data to struct. \n" + err.Error())
		return
	}

	var message string
	message = firebase.Weather

	jsonMessage := structs.NotificationResponse{
		Text: message,
	}

	//Updating the new time, from weather conditions
	timeS, err := time.Parse(time.RFC822, firebase.ArrivalTime)
	if err != nil {
		log.Println("Error when parsing time in send notification " + err.Error())
	}
	newTime = timeS.Add(time.Duration(-firebase.EstimatedTravelTime) * time.Minute)
	timeString := newTime.Add(time.Minute).String()

	text := "Your registered trip is about to begin. To be there in time, consider departure " + timeString + "\n\n\n" +
		jsonMessage.Text + ". For more information go to our website:"

	//link for the weather endpoint
	link := "http://10.212.141.222:80/rtc/v1/weather/" + destination

	// Formats output to be accepted by Slack
	jsonData := structs.JsonMessage{Text: text, Attachment: []structs.Attachments{{
		Color:      "#2eb886",
		AuthorName: "Roadtrip Planner",
		Title:      "Weather",
		TitleLink:  link,
		Text:       "The Weather Forecast for you destination",
		Footer:     "The Road trip Companion",
	}}}

	output, err := json.Marshal(jsonData)
	if err != nil {
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		log.Println(jsonError)
		return
	}

	// Checks if the user has deleted the webhook while the webhook was asleep
	// if so -> the function returns and the user will not get notified.
	_, err = database.Get(notificationId)
	if err != nil {
		log.Println("Unable to find webhook with ID: " + notificationId)
		return
	}

	// Creates a go routine of the invocation
	go CallUrl(notificationUrl, string(output))
}

// InvokeAll Invokes all webhooks
func InvokeAll() {
	webhook, err := database.GetAll()
	if err != nil {
		log.Println("There has been an error retrieving the webhooks from the database.\n" + err.Error())
		log.Fatalf(err.Error())
		return
	}
	// For each webhook, create a go routine for it
	for i := 0; i < len(webhook); i++ {
		go SendNotification(webhook[i].Ref.ID)
	}
}
