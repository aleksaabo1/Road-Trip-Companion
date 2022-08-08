package utils

import (
	"log"
	"time"
)

// IsValidInput Checks if the arrival time input is valid for the webhook registration.
// Needs to follow the RFC822 format.
func IsValidInput(arrivalTime string) bool {
	if len(arrivalTime) != 0 {
		// Tries to parse the arrivalTime to the RFC822-format
		_, err := time.Parse(time.RFC822, arrivalTime)
		if err != nil {
			return false
		}
		return true
	} else {
		log.Println("There was an error parsing the arrival time: " + arrivalTime + " to the RFC822-format.")
		return false
	}
}
