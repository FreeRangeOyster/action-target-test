package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})

	fmt.Println("Server is running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
