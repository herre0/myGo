package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("server is running");
	http.HandleFunc("/", func(rw http.ResponseWriter, res *http.Request) {
        switch res.Method {
			case "GET":
			rw.Write([]byte("GET/ Hello World"))
			case "POST":
			rw.Write([]byte("POST/ Hello World"))
		}
    })

    log.Fatal(http.ListenAndServe(":8000", nil))
}