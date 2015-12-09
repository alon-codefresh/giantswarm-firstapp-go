package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/garyburd/redigo/redis"
)

var redisCon redis.Conn

const KelvinToCelsiusDiff = 273
const OpenWeatherMapAPIKey = "182564eaf55f709a58a13c40086fb5bb"

type WeatherReport struct {
	Main struct {
		Temperature float64 `json:"temp"`
	}
	Sys struct {
		Country string `json:"country"`
	}
	Name  string `json:"name"`
	Error string `json:"message"`
}

func main() {
	var err error
	log.Println("Establishing connection to Redis")
	redisCon, err = redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Fatalf("Could not connect to Redis with error: %s", err)
	}
	defer redisCon.Close()

	http.HandleFunc("/", currentWeatherHandler)

	go func() {
		log.Println("Starting current weather server at :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
}

func currentWeatherHandler(w http.ResponseWriter, r *http.Request) {
	report, err := getWeatherReport(r.URL.Query().Get("q"))
	if err != nil {
		fmt.Fprintf(w, "Cannot get weather data: %s\n", err)
	} else if len(report.Error) > 1 {
		fmt.Fprintf(w, "%s\n", report.Error)
	} else {
		celsius := report.Main.Temperature - KelvinToCelsiusDiff
		fmt.Fprintf(w, "Current temperature in %v (%v) is %.1f °C\n", report.Name, report.Sys.Country, celsius)
	}
}

func getWeatherReport(query string) (WeatherReport, error) {
	var report WeatherReport

	data, err := cacheReport(getWeatherReportData, query)
	if err != nil {
		return report, err
	}

	if err = json.Unmarshal(data, &report); err != nil {
		return report, err
	}

	return report, nil
}

func getWeatherReportData(query string) ([]byte, error) {
	var data []byte

	if query == "" {
		query = "Cologne,DE"
	}

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%v&appid=%v", url.QueryEscape(query), OpenWeatherMapAPIKey)

	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	return data, nil
}

func cacheReport(f func(string) ([]byte, error), param string) ([]byte, error) {
	key := fmt.Sprintf("report_%x", md5.Sum([]byte(param)))
	data, _ := redis.Bytes(redisCon.Do("GET", key))
	if len(data) == 0 {
		log.Println("Querying live weather data")
		res, err := f(param)
		if err != nil {
			return nil, err
		}
		redisCon.Do("SETEX", key, 60, res)
		data = res
	} else {
		log.Println("Using cached weather data")
	}
	return data, nil
}
