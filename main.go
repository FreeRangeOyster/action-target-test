package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	hostListPtr := flag.String("hosts", "", "A space-delimited list of hosts to monitor.")
	portPtr := flag.Int("port", 80, "The port to monitor on each host. Defaults to 80.")
	intervalPtr := flag.Int("interval", 5000, "The interval on which to check each host in milliseconds. Defaults to 5000.")
	flag.Parse()

	fmt.Println(*hostListPtr)
	trimmedList := strings.TrimSpace(*hostListPtr)
	if trimmedList == "" {
		panic("No hosts provided")
	}
	fmt.Println("Preparing to monitor port ", *portPtr, " every ", *intervalPtr, " milliseconds on hosts", *hostListPtr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})

	fmt.Println("Server is running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
