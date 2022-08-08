package webhooks

import (
	"bytes"
	"cloudproject/database"
	"cloudproject/structs"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateWebhook(t *testing.T) {
	database.Ctx = context.Background()
	sa := option.WithCredentialsFile("cloudprojecttwo-firebase-adminsdk-uke12-6ed6b4ca4e.json")
	app, err := firebase.NewApp(database.Ctx, nil, sa)
	if err != nil {
		_ = fmt.Errorf("error initializing app: %v", err)
	}

	database.Client, err = app.Firestore(database.Ctx)
	if err != nil {
		log.Fatalln(err)
	}

	mcPostBody := map[string]interface{}{
		"url":                "https://discord.com/api/webhooks/842330664279998474/YpO-9WUDl9qwl29ka9wvlm90ijN_gZeYkWwIfJl41IXRNUWYH3EMDH6hWBeZbbHKwDSz",
		"ArrivalDestination": "lillehammer",
		"DepartureLocation":  "gjøvik",
		"ArrivalTime":        "10 aug 21 12:10 CEST",
	}
	body, _ := json.Marshal(mcPostBody)

	req, err := http.NewRequest("POST", "http://localhost:8080/hook/", bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatalf("could not create request %v", err)
	}

	rec := httptest.NewRecorder()
	AddWebhook(rec, req)

	response := rec.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Errorf("Expected status Ok; got %v", response.StatusCode)
	}

	read, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Got an error in reading")
	}

	output := string(read)
	outputarr := strings.Split(output, "{")
	id := strings.Split(outputarr[0], ":")[1]

	webhook, err := database.Get(strings.TrimSpace(id))

	var hook structs.Webhook
	err = json.Unmarshal(webhook, &hook)
	if err != nil {
		t.Fatalf("Error when unmashaling")
	}

	if "lillehammer" != hook.ArrivalDestination {
		t.Fatalf("Expected lillehammer; got %v", hook.ArrivalDestination)
	} else if "gjøvik" != hook.DepartureLocation {
		t.Fatalf("Expected gjøvik; got %v", hook.DepartureLocation)
	}
	_, err = database.Client.Collection(database.Collection).Doc(strings.TrimSpace(id)).Delete(database.Ctx)
	if err != nil {
		t.Fatalf("Error when deleting")
	}
}
