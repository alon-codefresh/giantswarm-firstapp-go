package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("test")
	for i := 0; i < 100; i++ {
		for i := 0; i < 10; i++ {
			go ping()
		}
		time.Sleep(1 * time.Second)
	}
}

func ping() {
	fmt.Println("ping")
	resp, err := http.Get("http://192.168.59.103:8080")
	if err != nil {
		log.Printf("error %s\n", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

}
