package main

import (
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
)

func main() {
	c := resty.New()
	srtURL := "http://localhost:8080/sub?topic=test&id=testid&timeout=10"

	for {
		resp, err := c.R().Get(srtURL)
		if err != nil {
			log.Printf("HTTP request err:%s", err)
		}

		if resp.StatusCode() != http.StatusOK {
			log.Printf("HTTP status %s", resp.Status())
		}

		log.Printf("%s", resp.Body())
	}

}
