package webhooks

import (
	"cloud.google.com/go/firestore"
	"cloudproject/database"
	"cloudproject/endpoints"
	"cloudproject/structs"
	"cloudproject/utils"
	"encoding/json"
	"errors"
	"fmt"
	_ "fmt"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
	"strings"
	"sync"
	"time"
	_ "time"
)

// Check Checks for updates in weather conditions and traffic incidents
func Check(w http.ResponseWriter) {
	// Loop through all entries in collection "messages"
	iter := database.Client.Collection(database.Collection).Documents(database.Ctx)
	var hook structs.Webhook

	// While loop
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		// Adds the data to the Webhook-struct
		if err := doc.DataTo(&hook); err != nil {
			log.Println("There was an error while adding data to the struct.\n" + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		weatherMessage := hook.Weather

		// Receives the latitude and longitude of the place passed in to the url
		latitude, longitude, err := database.LocationPresent(url2.QueryEscape(hook.DepartureLocation))
		if err != nil {
			log.Println("There was an error while retrieving GeoCode for location: " + hook.DepartureLocation + "\n" + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		url := ""
		if latitude != "" && longitude != "" {
			// Defines the url to the openweathermap API with relevant latitude and longitude and apiKey
			url = "https://api.openweathermap.org/data/2.5/weather?lat=" + latitude + "&lon=" + longitude + "&appid=" + utils.OpenweathermapKey
		} else {
			log.Println("Wrong formatting or no input for latitude and/or longitude.")
			http.Error(w, "Check formatting of latitude and longitude.", http.StatusBadRequest)
		}

		// Gets the current weather
		newMessage := endpoints.CurrentWeatherHandler(w, url).Main.Message
		if !(newMessage == weatherMessage) {
			hook.Weather = newMessage
			database.Update(doc.Ref.ID, hook)
		}
	}
	time.Sleep(time.Minute * 30)
	Check(w)
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Split(r.URL.Path, "/")[4]
	switch r.Method {
	case http.MethodDelete:
		message, err := database.Delete(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			_, err := fmt.Fprintf(w, message)
			if err != nil {
				log.Println(err.Error())
			}

		}
	case http.MethodPost:
		AddWebhook(w, r)
	case http.MethodGet:
		w.Header().Set("Content-type", "application/json")
		if id != "" {
			output, _ := database.Get(id)
			_, err := fmt.Fprintf(w, "%v", string(output))
			if err != nil {
				log.Println(err.Error())
			}
		} else {
			GetAllWebhooks(w)
		}
	}
}

// AddWebhook Add new webhook
func AddWebhook(w http.ResponseWriter, r *http.Request) {

	// Need to wait before starting new go routine
	wg := new(sync.WaitGroup)
	// Read response body
	input, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("There was an error during read of response body.\n" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(input) == 0 {
		log.Println("The message is empty.")
		http.Error(w, "Your message appears to be empty", http.StatusBadRequest)
		return
	}

	var notification structs.Webhook
	// Unmarshalling the body
	if err = json.Unmarshal(input, &notification); err != nil {
		log.Println("There was an error during unmarshalling.\n" + err.Error())
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

	// Checks the webhook format
	err = webhookFormat(notification)
	if err != nil {
		log.Println("Error: Check webhook format.\n" + err.Error())
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	// Adds data to the database
	id, _, err := database.Client.Collection(database.Collection).Add(database.Ctx,
		map[string]interface{}{
			"url":                notification.Url,
			"ArrivalDestination": notification.ArrivalDestination,
			"ArrivalTime":        notification.ArrivalTime,
			"Weather":            notification.Weather,
			"DepartureLocation":  notification.DepartureLocation,
		})
	if err != nil {
		log.Println("Error: Unable to add data to database.\n" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		trimmedId := strings.TrimLeft(id.ID, "/") //Trimming the id
		//Adding the Id to a field, for easier access
		_, err = database.Client.Collection(database.Collection).Doc(trimmedId).Set(database.Ctx, map[string]interface{}{
			"id": trimmedId,
		}, firestore.MergeAll)
		log.Println("Successfully registered webhook with ID: " + id.ID)
		http.Error(w, "Registered with ID: "+id.ID, http.StatusCreated)
		go Check(w)

		// Waits for Check to complete before moving on
		wg.Wait()
		err := CalculateDeparture(id.ID)
		if err != nil {
			database.Delete(id.ID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go SendNotification(id.ID)
	}
}

// webhookFormat Checks the format of the webhook
func webhookFormat(web structs.Webhook) error {

	if web.DepartureLocation == "" {
		log.Println("Departure location cannot be empty.")
		return errors.New("error, departure location cannot be empty")
	} else if web.ArrivalDestination == "" {
		log.Println("Arrival destination cannot be empty.")
		return errors.New("error, arrival destination cannot be empty")
	} else if web.ArrivalTime == "" {
		log.Println("Arrival time cannot be empty.")
		return errors.New("error, arrival time cannot be empty")
	}
	err := utils.IsValidInput(web.ArrivalTime)
	if !err {
		log.Println("Error: Invalid time format. Example of expected format: 17 may 21 12:10 CEST")
		return errors.New("Invalid time format\nExample of expected format: 17 may 21 12:10 CEST")
	}
	return nil
}

// DeleteExpiredWebhooks Deletes webhooks which are older than 24 hours
func DeleteExpiredWebhooks() {
	// Retrieves all entries in collection "messages"
	iter := database.Client.Collection(database.Collection).Documents(database.Ctx)

	var firebase structs.Webhook

	// Iterates through all instances
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err := doc.DataTo(&firebase); err != nil {
			log.Println("Unable to append data to firebase.")
		}

		arrival, err := time.Parse(time.RFC822, firebase.ArrivalTime)
		if err != nil {
			log.Println("Unable to parse time to format: RFC822.")
			log.Fatalf(err.Error())
		}

		if arrival.Before(time.Now().AddDate(0, 0, -1)) {
			_, err := database.Delete(doc.Ref.ID)
			if err != nil {
				log.Println("Deletion of webhook with ID: " + doc.Ref.ID + " FAILED.")
				log.Fatalf(err.Error())
			}
			log.Println("Webhook got SUCCESSFULLY deleted.")
		}
	}
	time.Sleep(time.Hour * 24)
	DeleteExpiredWebhooks()
}

func GetAllWebhooks(w http.ResponseWriter) {
	w.Header().Set("Content-type", "application/json")

	list, err := database.GetAll()
	if err != nil {
		log.Println("Error: Error encountered while retrieving data.")
		http.Error(w, "Error occurred when listing webhooks from database", http.StatusInternalServerError)
		return
	}

	var allWebhooks []structs.Webhook
	for _, doc := range list {
		webhook := structs.Webhook{}
		if err := doc.DataTo(&webhook); err != nil {
			http.Error(w, "Tried to iterate through webhooks but failed", http.StatusBadRequest)
		}
		webStruct := structs.Webhook{
			Id:                  webhook.Id,
			Url:                 webhook.Url,
			DepartureLocation:   webhook.DepartureLocation,
			ArrivalDestination:  webhook.ArrivalDestination,
			Weather:             webhook.Weather,
			ArrivalTime:         webhook.ArrivalTime,
			EstimatedTravelTime: webhook.EstimatedTravelTime,
		}
		allWebhooks = append(allWebhooks, webStruct)
	}

	// Marshalling the array to JSON
	output, err := json.Marshal(allWebhooks)
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
