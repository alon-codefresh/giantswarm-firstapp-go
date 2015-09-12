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
	"time"

	"github.com/garyburd/redigo/redis"
)

var pool *redis.Pool

const KelvinToCelsiusDiff = 273

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
	pool = newPool()

	if err != nil {
		log.Fatalf("Could not connect to Redis with error: %s", err)
	}

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

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + url.QueryEscape(query))
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

	redisCon := pool.Get()
	defer redisCon.Close()

	data, _ := redis.Bytes(redisCon.Do("GET", key)) // err is a cache miss

	if len(data) == 0 {
		start := time.Now()
		res, err := f(param)
		log.Printf("Queried live weather data in %s\n", time.Since(start))
		if err != nil {
			return nil, err
		}

		reply, err := redisCon.Do("SETEX", key, 30, res)
		if reply != nil {
			log.Printf("redisCon.Do(SETEX) reply: %s\n", reply)
		}
		if err != nil {
			return data, err
		}

		data = res
	} else {
		log.Println("Using cached weather data")
	}
	return data, nil
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddress())
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
}

func redisAddress() string {
	addr := os.Getenv("REDIS_PORT_6379_TCP_ADDR")
	port := os.Getenv("REDIS_PORT_6379_TCP_PORT")
	return net.JoinHostPort(addr, port)
}
