package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HostStatus string

const (
	HostStatusOnline   HostStatus = "Online"
	HostStatusUnstable HostStatus = "Unstable"
	HostStatusOffline  HostStatus = "Offline"
)

type host struct {
	Hostname                 string
	Status                   HostStatus
	LastSeen                 time.Time
	SessionAverageLatency    uint16
	FiveMinuteAverageLatency uint16
	FiveMinuteFailures       uint64
}

type dashboardData struct {
	SessionStart time.Time
	Port         int
	Hosts        map[string]host
}

func main() {
	hostListPtr := flag.String("hosts", "", "A space-delimited list of hosts to monitor.")
	portPtr := flag.Int("port", 80, "The port to monitor on each host. Defaults to 80.")
	intervalPtr := flag.Int("interval", 5000, "The interval on which to check each host in milliseconds. Minimum 1000. Defaults to 5000.")
	flag.Parse()

	trimmedList := strings.TrimSpace(*hostListPtr)
	if trimmedList == "" {
		panic("No hosts provided")
	}
	if *portPtr <= 0 || *portPtr > 65535 {
		panic("Invalid port provided")
	}
	if *intervalPtr <= 1000 {
		panic("Invalid interval provided")
	}
	interval := time.Duration(*intervalPtr * 1000 * 1000)

	fmt.Println("Preparing to monitor port", *portPtr, "every", *intervalPtr, "milliseconds on hosts", *hostListPtr)

	tmpl := template.Must(template.New("dashboard.tmpl").ParseFiles("dashboard.tmpl"))

	data := dashboardData{
		SessionStart: time.Now(),
		Port:         *portPtr,
		Hosts:        make(map[string]host),
	}
	for hostname := range strings.SplitSeq(*hostListPtr, " ") {
		data.Hosts[hostname] = host{
			Hostname:                 hostname,
			SessionAverageLatency:    0,
			FiveMinuteAverageLatency: 0,
			FiveMinuteFailures:       0,
		}
		go checkHost(hostname, *portPtr, interval)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
		}
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
