package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	port = ":8083"
)

func main() {
	http.HandleFunc("/test", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":
			defer request.Body.Close()
			var body, _ = ioutil.ReadAll(request.Body)
			fmt.Println(request, string(body))
			fmt.Println("------")
		default:
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, "only POST method allowed")
		}
	})

	log.Println("test server up at ", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
