package main

import (
	"fmt"
	"net/http"

	"github.com/mi-wada/retryrt"
)

func main() {
	client := &http.Client{}
	client.Transport = retryrt.New(
		client.Transport,
	)

	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	resp.Body.Close()
}
