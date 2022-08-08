package endpoints

import (
	"cloudproject/database"
	"cloudproject/structs"
	"cloudproject/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CurrentWeather Gets location and passes it to the openweathermap-API
func CurrentWeather(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-type", "application/json")

	// Splits the URL to get the name of the city to be checked
	address := strings.Split(request.URL.Path, `/`)[4]

	//Receives the latitude and longitude of the place passed in the url
	latitude, longitude, err := database.LocationPresent(url.QueryEscape(address))
	if err != nil {
		log.Println("Error: No entries in the database gave the wanted result.\n" + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	urlLoc := ""

	if latitude != "" && longitude != "" {
		// Defines the url to the openweathermap API with relevant latitude and longitude and apiKey
		urlLoc = "https://api.openweathermap.org/data/2.5/weather?lat=" + latitude + "&lon=" + longitude + "&appid=" + utils.OpenweathermapKey
	} else {
		_, err := fmt.Fprint(rw, "Check formatting of latitude and longitude.")
		if err != nil {
			log.Println("Error: Error while displaying data to the user.\n" + err.Error())
			http.Error(rw, "Error encountered while attempting to display data.", http.StatusInternalServerError)
			return
		}
	}
	// Calls the handler using the URL.
	test := CurrentWeatherHandler(rw, urlLoc)
	// Marshal the struct
	output, err := json.Marshal(test) //Marshalling the array to JSON
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	// Print the weather to the user in json format
	_, err = fmt.Fprintf(rw, "%v", string(output))
	if err != nil {
		log.Println("There was an error while displaying the output to the user.")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} //Outputs the weather
}

// CurrentWeatherHandler Handling request with the url
func CurrentWeatherHandler(rw http.ResponseWriter, url string) structs.OutputWeather {
	// Uses request URL
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error: Encountered problem when requesting the url.\n" + err.Error())
		http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
		return structs.OutputWeather{}
	}

	// Reads the data from the resp.Body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading response body.\n" + err.Error())
		http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
		return structs.OutputWeather{}
	}

	// Defines struct instance
	var weather structs.WeatherData

	// Unmarshalling the body into the weatherData struct/fields
	if err = json.Unmarshal(body, &weather); err != nil {
		log.Println("There was an error during unmarshalling.\n" + err.Error())
		jsonError := utils.JsonUnmarshalErrorHandling(err)
		http.Error(rw, jsonError.Error(), http.StatusInternalServerError)
		return structs.OutputWeather{}
	}

	// Defines output struct instance
	var data []structs.OutputWeather

	// Defines various temporary variables with the data from the struct
	main := weather.Weather[0].Main
	rain1H := weather.Rain.OneH
	snow1H := weather.Snow.OneH
	tempActual := math.Round((weather.Main.Temp-273.15)*100) / 100
	tempFeelsLike := math.Round((weather.Main.FeelsLike-273.15)*100) / 100
	tempMin := math.Round((weather.Main.TempMin-273.15)*100) / 100
	tempMax := math.Round((weather.Main.TempMax-273.15)*100) / 100
	humidity := weather.Main.Humidity
	visibility := weather.Visibility
	windSpeed := weather.Wind.Speed
	windDeg := weather.Wind.Deg
	sunrise := weather.Sys.Sunrise
	sunset := weather.Sys.Sunset

	// Attaches the various temporary variables to the output struct
	jsonStruct := structs.OutputWeather{
		Main:       structs.MainStruct{Main: main},
		Rain1h:     rain1H,
		Snow1h:     snow1H,
		Temp:       structs.TempStruct{Temp: tempActual},
		FeelsLike:  structs.FeelsLikeStruct{FeelsLike: tempFeelsLike},
		TempMin:    structs.TempMinStruct{TempMin: tempMin},
		TempMax:    structs.TempMaxStruct{TempMax: tempMax},
		Humidity:   structs.HumidityStruct{Humidity: humidity},
		Visibility: structs.VisibilityStruct{Visibility: visibility},
		WindSpeed:  structs.WindSpeedStruct{WindSpeed: windSpeed},
		WindDeg:    structs.WindDegStruct{WindDeg: windDeg},
		Sunrise:    structs.SunriseStruct{Sunrise: sunrise},
		Sunset:     structs.SunsetStruct{Sunset: sunset}}

	// Appends the struct to the array
	data = append(data, jsonStruct)

	// Calls method response which returns an array containing different return messages
	responseArr := response(rw, data)

	// Redefines jsonStruct to also contain the different, relevant messages
	jsonStruct = structs.OutputWeather{
		Main:       structs.MainStruct{Main: main, Message: responseArr[0]},
		Rain1h:     rain1H,
		Snow1h:     snow1H,
		Temp:       structs.TempStruct{Temp: tempActual, Message: responseArr[1]},
		FeelsLike:  structs.FeelsLikeStruct{FeelsLike: tempFeelsLike, Message: responseArr[2]},
		TempMin:    structs.TempMinStruct{TempMin: tempMin, Message: responseArr[3]},
		TempMax:    structs.TempMaxStruct{TempMax: tempMax, Message: responseArr[4]},
		Humidity:   structs.HumidityStruct{Humidity: humidity, Message: responseArr[5]},
		Visibility: structs.VisibilityStruct{Visibility: visibility, Message: responseArr[6]},
		WindSpeed:  structs.WindSpeedStruct{WindSpeed: windSpeed, Message: responseArr[7]},
		WindDeg:    structs.WindDegStruct{WindDeg: windDeg, Message: responseArr[8]},
		Sunrise:    structs.SunriseStruct{Sunrise: sunrise, Message: responseArr[9]},
		Sunset:     structs.SunsetStruct{Sunset: sunset, Message: responseArr[10]}}

	return jsonStruct
}

// response Handles the different response-messages depending on the weather conditions
// returns an array containing all the various return messages
func response(rw http.ResponseWriter, data []structs.OutputWeather) []string {

	// Defines the different messages as string
	var mainMessage string
	var tempMessage string
	var feelsLikeMessage string
	var tempMinMessage string
	var tempMaxMessage string
	var humidityMessage string
	var visibilityMessage string
	var windSpeedMessage string
	var windDegMessage string
	var sunriseMessage string
	var sunsetMessage string

	// Defines the struct data passed to the method as dataTemp
	dataTemp := data[0]

	// Switch-cases handling different messages for different weather conditions
	// Switch on Main weather condition
	switch dataTemp.Main.Main {
	// It is raining
	case "Rain":
		// Switch while true, to have a check for the different variables
		switch true {
		// It has rained less than or equal to 2 mm the last hour -> this is the message to be returned -> break out of switch true loop
		case dataTemp.Rain1h <= 2:
			mainMessage = "It has been light rain the last hour, consider bringing a umbrella."
			break
		// It has rained more than 2 mm and less than or equal to 7 mm the last hour -> this is the message to be returned -> break out of switch true loop
		case dataTemp.Rain1h > 2 && dataTemp.Rain1h <= 7:
			mainMessage = "It has been moderate rain the last hour, consider bringing rainwear."
			break
		case dataTemp.Rain1h > 7 && dataTemp.Rain1h <= 50:
			mainMessage = "It has been heavy rain the last hour, you should bring rainwear."
			break
		case dataTemp.Rain1h > 50:
			mainMessage = "It has been violent rain the last hour, you will most likely become wet if you go outside."
			break
		// None of the cases above is relevant : have a default return message
		default:
			mainMessage = "It is raining, bring appropriate clothing."
			break
		}
	// It is snowing
	case "Snow":
		// Switch while true, to have a check for the different variables
		switch true {
		// It has snowed less than or equal to 1 mm the last hour -> this is the message to be returned -> break out of switch true loop
		case dataTemp.Snow1h <= 1:
			mainMessage = "It has been snowing lightly the last hour, consider winter tires, " +
				"wear appropriate clothing and set of a couple of minutes to clear the snow of your car."
			break
		// It has rained more than 1 mm and less than or equal to 3 mm the last hour -> this is the message to be returned -> break out of switch true loop
		case dataTemp.Snow1h > 1 && dataTemp.Snow1h <= 3:
			mainMessage = "It has been snowing moderately the last hour, make sure to have winter tires, " +
				"wear appropriate clothing and expect at least 10 minutes to clear the snow of your car."
			break
		case dataTemp.Snow1h > 3:
			mainMessage = "It has been snowing heavily the last hour, you must have winter tires, " +
				"wear appropriate clothing and expect at least 20 minutes to clear the snow around- and of your car."
			break
		// None of the cases above is relevant : have a default return message
		default:
			mainMessage = "It is snowing, drive carefully, bring appropriate clothing and turn on the heater."
			break
		}
	// It is clear sky outside
	case "Clear":
		// No need for switch true loop, because nothing more needs to be checked
		// This is the return message
		mainMessage = "The sky is clear, wear appropriate clothing with respect to terrain and temperature."
	default:
		mainMessage = "The weather of your destination is: " + dataTemp.Main.Main
	}

	// Reformatting temp to one decimal
	s := fmt.Sprintf("%.1f", dataTemp.Temp.Temp)
	switch true {
	case dataTemp.Temp.Temp <= 0:
		tempMessage = "The temperature is: " + s + " degrees Celsius. It is below the freezing point outside, drive carefully " +
			"and we recommend you to use winter tires. Bring warm clothes and something warm to drink."
	case dataTemp.Temp.Temp > 0 && dataTemp.Temp.Temp <= 10:
		tempMessage = "The temperature is: " + s + " degrees Celsius. The temperature outside is moderate. Consider whether " +
			"winter tires is needed. Most likely summer tires are recommended. Wear a moderate amount of clothing."
		break
	case dataTemp.Temp.Temp > 10 && dataTemp.Temp.Temp < 20:
		tempMessage = "The temperature is: " + s + " degrees Celsius. The temperature outside is relatively high. Summer tires " +
			"are needed. Wear a moderate amount of clothing."
		break
	case dataTemp.Temp.Temp >= 20:
		// If the temp is above 20 degrees celsius -> Check whether it is clear sky or not
		switch true {
		// Sky is clear -> This is the return message -> break out of the switch true loop
		case dataTemp.Main.Main == "Clear":
			tempMessage = "The temperature is: " + s + " degrees Celsius. The temperature outside is high, and summer tires are highly " +
				"recommended! Put on some sunscreen, some light clothing, blast your air condition at max and put a slush in the " +
				"cup holder. Enjoy the temperature!"
			break
		// Break out of the loop if the weather is not clear
		default:
			break
		}
		// This is the message to be used
		tempMessage = "The temperature is: " + s + " degrees Celsius. The temperature outside is high, and summer tires are highly " +
			"recommended! Consider the need for sunscreen, appropriate clothing and drive after the conditions."
		break
	default:
		tempMessage = "The temperature is: " + s + " degrees Celsius. Consider whether winter tires is needed and wear appropriate clothing."
		break
	}

	t := fmt.Sprintf("%.1f", dataTemp.FeelsLike.FeelsLike)
	switch true {
	case dataTemp.FeelsLike.FeelsLike <= 10:
		feelsLikeMessage = "It feels cold outside, with a feels like temperature of: " + t + " degrees Celsius."
		break
	case dataTemp.FeelsLike.FeelsLike > 10 && dataTemp.FeelsLike.FeelsLike <= 20:
		feelsLikeMessage = "It is a great temperature outside for jeans and jumper, with a feels like temperature of: " + t + " degrees Celsius."
		break
	case dataTemp.FeelsLike.FeelsLike > 20:
		switch true {
		case dataTemp.Main.Main == "Clear":
			feelsLikeMessage = "It is warm outside today, light clothes are strongly recommended. Take a trip to the beach " +
				"and enjoy the day. The temperature feels like: " + t + " degrees Celsius."
			break
		default:
			break
		}
		feelsLikeMessage = "It is warm outside today, consider to use light clothes unless it is raining. " +
			"The temperature feels like: " + t + " degrees Celsius."
		break
	default:
		feelsLikeMessage = "The temperature outside feels like: " + t + " degrees Celsius. Dress accordingly."
		break
	}

	u := fmt.Sprintf("%.1f", dataTemp.TempMin.TempMin)
	tempMinMessage = "The minimum temperature for the day is expected to be: " + u + " degrees Celsius. Consider this both with regards " +
		"to tires used and clothes you plan to wear."

	v := fmt.Sprintf("%.1f", dataTemp.TempMax.TempMax)
	tempMaxMessage = "The maximum temperature for the day is expected to be: " + v + " degrees Celsius. Consider this both with regards " +
		"to tires used and clothes you plan to wear."

	switch true {
	case dataTemp.Humidity.Humidity <= 30:
		humidityMessage = "The humidity value at the moment is:" + strconv.Itoa(dataTemp.Humidity.Humidity) + " percent. " +
			"This is relatively low humidity which can result in health issues."
		break
	case dataTemp.Humidity.Humidity > 30 && dataTemp.Humidity.Humidity <= 75:
		humidityMessage = "The humidity values at the moment is: " + strconv.Itoa(dataTemp.Humidity.Humidity) + " percent. " +
			"This is an ideal humidity value and you should be comfortable going outside."
		break
	case dataTemp.Humidity.Humidity > 75:
		humidityMessage = "The humidity values at the moment is: " + strconv.Itoa(dataTemp.Humidity.Humidity) + " percent. " +
			"If it ain't raining already, there is a high change it will start raining soon."
		break
	default:
		humidityMessage = "The humidity values at the moment is: " + strconv.Itoa(dataTemp.Humidity.Humidity) + " percent."
		break
	}

	visibilityMessage = "The visibility outside is: " + strconv.Itoa(dataTemp.Visibility.Visibility) + " meters."

	// Rounds of wind speed to 2 decimals
	x := fmt.Sprintf("%.2f", dataTemp.WindSpeed.WindSpeed)
	windSpeedMessage = "The wind speed today is " + x + " m/s."

	windDegMessage = "The wind degree direction is: " + strconv.Itoa(dataTemp.WindDeg.WindDeg)

	// Convert from epoch to human readable date
	// Inspired by/taken from:
	//	- https://play.golang.org/p/6h0A0WPxtq
	//	- https://www.epochconverter.com
	sunriseEpoch := dataTemp.Sunrise.Sunrise
	sunsetEpoch := dataTemp.Sunset.Sunset

	// Returns message with the time for sunrise in a human readable format
	sunriseMessage = "The time for sunrise is: " + epochToHumanReadable(int64(sunriseEpoch)).String()

	// Returns message with the time for sunrise in a human readable format
	sunsetMessage = "The time for sunset is: " + epochToHumanReadable(int64(sunsetEpoch)).String()

	// Manually applies the different messages to an array ready to be returned
	messages := []string{mainMessage, tempMessage, feelsLikeMessage, tempMinMessage, tempMaxMessage, humidityMessage,
		visibilityMessage, windSpeedMessage, windDegMessage, sunriseMessage, sunsetMessage}

	// Array to be returned
	var returnedMessages []string

	// Loop variable
	var i int

	// For each message in the messages array -> append them to the returnedMessages array
	for i = 0; i < len(messages); i++ {
		returnedMessages = append(returnedMessages, messages[i])
	}

	// Return all the messages as an array of messages
	return returnedMessages
}

// epochToHumanReadable Converts time from epoch format to a more human readable RFC822-format
// Inspired by and taken from:
// - https://play.golang.org/p/6h0A0WPxtq
// - https://www.epochconverter.com
// We found this useful to use because we got the data from the API in Epoch and Unit format and this is not readable
// for the user, when the format is for instance: "1619838624". After conversion we get the following format:
// "2021-05-01 05: 10: 24 +0200 CEST"
//
func epochToHumanReadable(epoch int64) time.Time {
	return time.Unix(epoch, 0)
}

// GetMessageWeight
// Analyzes the weather message and calculates how much time needs to be added
// Based on weather conditions
func GetMessageWeight(message string) int {
	base := 10.0
	skyWeight := 1.0

	// Changes the weight of extra time needed based on weather
	switch true {
	case strings.Contains(message, "snow"):
		skyWeight = 1.6
		if strings.Contains(message, "light") {
			skyWeight = 1.2
		} else if strings.Contains(message, "moderate") {
			skyWeight = 1.6
		} else if strings.Contains(message, "heavily") {
			skyWeight = 2.0
		}
		base = base * skyWeight
		break

	case strings.Contains(message, "rain"):
		skyWeight = 1.2
		if strings.Contains(message, "light") {
			skyWeight = 1
		} else if strings.Contains(message, "moderate") {
			skyWeight = 1.2
		} else if strings.Contains(message, "heavy") {
			skyWeight = 1.5
		} else if strings.Contains(message, "violent") {
			skyWeight = 1.8
		}
		base = base * skyWeight
		break
	}
	return int(base)
}
