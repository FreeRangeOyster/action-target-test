package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	hostListPtr := flag.String("hosts", "", "A space-delimited list of hosts to monitor.")
	portPtr := flag.Int("port", 80, "The port to monitor on each host. Defaults to 80.")
	intervalPtr := flag.Int("interval", 5000, "The interval on which to check each host in milliseconds. Defaults to 5000.")
	flag.Parse()

	trimmedList := strings.TrimSpace(*hostListPtr)
	if trimmedList == "" {
		panic("No hosts provided")
	}
	if *portPtr <= 0 || *portPtr > 65535 {
		panic("Invalid port provided")
	}
	if *intervalPtr <= 0 {
		panic("Invalid interval provided")
	}
	interval := time.Duration(*intervalPtr * 1000 * 1000)

	fmt.Println("Preparing to monitor port", *portPtr, "every", *intervalPtr, "milliseconds on hosts", *hostListPtr)

	for host := range strings.SplitSeq(*hostListPtr, " ") {
		go checkHost(host, *portPtr, interval)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})

	fmt.Println("Server is running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkHost(hostname string, port int, interval time.Duration) {
	for {
		portStr := strconv.Itoa(port)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostname, portStr), time.Second)
		if err != nil {
			fmt.Println("Connection error:", hostname, err)
		}
		if conn != nil {
			fmt.Println("Opened", net.JoinHostPort(hostname, portStr))
			conn.Close()
		}
		time.Sleep(interval)
	}
}
