package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	url         = "http://whois.turuapi.my.id/v1/whois/mass-lookup"
	contentType = "application/json"
	numRequests = 1000 // Jumlah total permintaan untuk diuji
	numWorkers  = 10   // Jumlah goroutine untuk mengirim permintaan secara bersamaan
)

type RequestBody struct {
	Domain []string `json:"domain"`
}

func main() {
	// Daftar domain untuk dikirim
	domains := []string{
		"turulabs.com", "google.com", "youtube.com", "facebook.com",
		"twitter.com", "instagram.com", "wikipedia.org", "amazon.com",
		"yahoo.com", "reddit.com", "linkedin.com", "netflix.com",
		"tiktok.com", "microsoft.com", "apple.com", "baidu.com",
		"qq.com", "taobao.com", "tmall.com", "jd.com",
		"aliexpress.com", "weibo.com", "live.com", "vk.com",
		"sohu.com", "bing.com", "ebay.com", "pinterest.com",
		"whatsapp.com", "twitch.tv", "wordpress.com", "paypal.com",
		"quora.com", "github.com", "cnn.com", "bbc.com",
		"nytimes.com", "imdb.com", "weather.com", "tumblr.com",
		"dropbox.com", "adobe.com", "hulu.com", "espn.com",
		"foxnews.com", "salesforce.com", "shopify.com", "spotify.com",
		"zoom.us", "slack.com", "airbnb.com",
	}

	requestBody := RequestBody{Domain: domains}
	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Post(url, contentType, bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("Error making request:", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Received non-200 response: %d\n", resp.StatusCode)
			}
		}()

		time.Sleep(time.Millisecond * 10) // Menambahkan jeda antar permintaan untuk menghindari overload
	}

	wg.Wait()
	elapsedTime := time.Since(startTime)
	fmt.Printf("Completed %d requests in %s\n", numRequests, elapsedTime)
}
