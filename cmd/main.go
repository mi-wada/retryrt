package main

import (
	"net/http"
	"time"

	"github.com/mi-wada/retryrt"
)

func main() {
	client := &http.Client{}
	client.Transport = retryrt.New(
		client.Transport,
		// Set maximum number of retry attempts
		retryrt.WithMaxRetries(5),
		// Customize backoff strategy
		retryrt.WithBackoff(retryrt.DefaultBackoff(1*time.Second, 30*time.Second)),
		// Customize retry conditions
		retryrt.WithShouldRetry(func(req *http.Request, resp *http.Response, err error) bool {
			// For example, retry on HTTP status code 499
			if resp != nil && resp.StatusCode == 499 {
				return true
			}
			return retryrt.DefaultShouldRetry(req, resp, err)
		}),
	)

	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
}
