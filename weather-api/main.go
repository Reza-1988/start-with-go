package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// This struct represents only the JSON fields we need from the API response.
// The `json:"..."` tags tell the JSON decoder which JSON key maps to which Go field.
type weatherResp struct {
	Name    string `json:"name"`
	Weather []struct {
		Description string `json:"description"` // weather condition text
		Icon        string `json:"icon"`        // icon ID (optional, not used here)
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`       // temperature in Celsius
		FeelsLike float64 `json:"feels_like"` // feels like temperature
		Humidity  int     `json:"humidity"`   // humidity percentage
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"` // wind speed
		Deg   int     `json:"deg"`   // wind direction
	} `json:"wind"`
}

func main() {

	// Loads the .env file and registers all key/value pairs as environment variables.
	// After this line, os.Getenv("KEY") can access values from .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Reads the API key previously loaded from the .env file.
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENWEATHERMAP_API_KEY is not set in the .env file")
		return
	}

	// fmt.Sprintf inserts `apiKey` into the string where %s appears.
	// This is cleaner than string concatenation.
	apiUrl := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=Tehran&appid=%s&units=metric&lang=en",
		apiKey,
	)

	// Sends the HTTP GET request to the weather API.
	response, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println("Error making the Api request:", err)
		return
	}

	// Ensures that the connection body is closed no matter how the function exits.
	// Prevents memory leaks.
	defer response.Body.Close()

	// Struct where decoded JSON will be stored
	var data weatherResp

	// json.NewDecoder reads JSON directly from response.Body (a stream),
	// and .Decode(&data) converts JSON → struct.
	//
	// Passing &data means we want the decoder to modify the original struct.
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		fmt.Println("JSON decode error:", err)
		return
	}

	// Description may not exist if API doesn't return a weather array,
	// so we define desc outside and only update it if the array has items.
	var desc string
	if len(data.Weather) > 0 {
		desc = data.Weather[0].Description // NOTE: "=" not ":=" (only assigning)
	}

	// "Humidity: %d%%"
	// %d → prints the integer humidity value
	// %% → prints a literal percent sign (%)
	fmt.Printf(
		"City: %s\nTemperature: %.1f°C\nFeels Like: %.1f°C\nHumidity: %d%%\nCondition: %s\n",
		data.Name, data.Main.Temp, data.Main.FeelsLike, data.Main.Humidity, desc,
	)
}
