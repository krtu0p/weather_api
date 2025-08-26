package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

type WeatherData struct {
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	CurrentWeather struct {
		Temperature   float64 `json:"temperature"`
		Windspeed     float64 `json:"windspeed"`
		Winddirection float64 `json:"winddirection"`
		Weathercode   int     `json:"weathercode"`
		Time          string  `json:"time"`
	} `json:"current_weather"`
}

func getWeather(lat, lon float64) (WeatherData, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true&temperature_unit=celsius&windspeed_unit=kmh&precipitation_unit=mm&timezone=auto", lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		return WeatherData{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WeatherData{}, err
	}

	var weather WeatherData
	err = json.Unmarshal(body, &weather)
	if err != nil {
		return WeatherData{}, err
	}

	return weather, nil

}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[2] == "" || pathParts[3] == "" {
		http.Error(w, "Latitude and Longitude required e.g /weather/51.5074/-0.1278", http.StatusBadRequest)
		return
	}

	lat, err := parseFloat(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	lon, err := parseFloat(pathParts[3])
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	weather, err := getWeather(lat, lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(weather)

}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func main() {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "http://127.0.0.1:5500"},
		AllowCredentials: true,
	})

	http.Handle("/", c.Handler(http.FileServer(http.Dir("."))))
	// CORS
	http.Handle("/weather/", c.Handler(http.HandlerFunc(weatherHandler)))

	fmt.Println("Starting server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
