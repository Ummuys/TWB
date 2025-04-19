package wapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type WeatherInfo struct {
	Location struct {
		Country string `json:"country"`
		Name    string `json:"name"`
	} `json:"location"`
	Curr struct {
		Temp          float32 `json:"temp_c"`
		Temp_feels    float32 `json:"feelslike_c"`
		Wind_kph      float32 `json:"wind_kph"`
		Cloud         float32 `json:"cloud"`
		Get_condition struct {
			Condition string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func PrintInfo(weather WeatherInfo) (string, string) {
	weatherInfo := fmt.Sprintf(
		`Страна: %s
Город: %s
Погода: %s
Температура: %.1f°C
Ощущается как: %.1f°C
Скорость ветра: %.1f км/ч
Процент облачности: %.1f`,
		weather.Location.Country,
		weather.Location.Name,
		weather.Curr.Get_condition.Condition,
		weather.Curr.Temp,
		weather.Curr.Temp_feels,
		weather.Curr.Wind_kph,
		weather.Curr.Cloud,
	)
	return weatherInfo, weather.Location.Name
}

func GetWeather(city string) (string, string, error) {
	location := url.QueryEscape(city)
	apiKey := os.Getenv("WEATHER_API")

	url := fmt.Sprintf(
		"http://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=%s",
		apiKey,
		location,
		"ru",
	)

	req, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("can't get url: %w", err)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", "", fmt.Errorf("problem with ReadAll: %w", err)
	}

	if strings.Contains(string(body), "error") {
		return "", "", nil
	}

	var weather WeatherInfo
	err = json.Unmarshal([]byte(body), &weather)
	if err != nil {
		return "", "", fmt.Errorf("can't Unmarshall: %w", err)
	}

	info1, info2 := PrintInfo(weather)

	return info1, info2, nil
}
