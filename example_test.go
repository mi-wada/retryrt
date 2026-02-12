package retryrt_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/mi-wada/retryrt"
)

func Example() {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "success")
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: retryrt.New(http.DefaultTransport),
	}

	resp, err := client.Get(srv.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
	// Output:
	// 200 OK
}
