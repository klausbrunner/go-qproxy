package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	port            = ":8082"
	putTimeoutMs    = 500
	outboundWorkers = 3
	queueSize       = 512
)

type ReceivedRequest struct {
	targetUrl string
	body      bytes.Buffer
	headers   http.Header
}

func (rr ReceivedRequest) String() string {
	return fmt.Sprintf("rr: %v", rr.targetUrl)
}

var requests = make(chan *ReceivedRequest, queueSize)

var httpClient = &http.Client{}

func processReceivedRequests(id int) {
	for {
		r := <-requests

		req, _ := http.NewRequest("POST", r.targetUrl, bytes.NewReader(r.body.Bytes()))
		req.Header = r.headers

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println(err)
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Println("request sent OK")
		} else {
			log.Printf("request failed: %v\n", err)
		}

		fmt.Printf("[%v] %v\n", id, r)
	}
}

func handleIncomingPost(response http.ResponseWriter, request *http.Request) {
	rr := new(ReceivedRequest)
	rr.headers = request.Header
	var reqUrl, err = url.Parse(request.RequestURI)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(response, "malformed request")
	}

	var targetUrl = reqUrl.Query().Get("url")
	if targetUrl == "" {
		response.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(response, "missing or malformed target url")
	}
	rr.targetUrl = targetUrl

	defer request.Body.Close()
	rr.body.ReadFrom(request.Body)

	select {
	case requests <- rr:
		response.WriteHeader(http.StatusAccepted)
	case <-time.After(putTimeoutMs * time.Millisecond):
		response.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(response, "queue full, slow down please")
	}

}

func main() {
	for i := 0; i < outboundWorkers; i++ {
		go processReceivedRequests(i)
	}

	http.HandleFunc("/qproxy", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":
			handleIncomingPost(response, request)
		default:
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, "only POST method allowed")
		}
	})

	log.Println("qproxy server up at ", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
