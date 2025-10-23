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

type Host struct {
	Hostname                 string
	Status                   HostStatus
	LastSeen                 time.Time
	SessionAverageLatency    uint16
	FiveMinuteAverageLatency uint16
	FiveMinuteFailures       uint64
}

type DashboardData struct {
	SessionStart time.Time
	Port         int
	Hosts        map[string]Host
}

type CheckLog struct {
	Timestamp time.Time
	Latency   uint16
}

type CheckLogMessage struct {
	Hostname string
	Log      CheckLog
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

	data := DashboardData{
		SessionStart: time.Now(),
		Port:         *portPtr,
		Hosts:        make(map[string]Host),
	}
	logs := make(map[string][]CheckLog)
	var channel chan CheckLogMessage = make(chan CheckLogMessage)
	for hostname := range strings.SplitSeq(*hostListPtr, " ") {
		data.Hosts[hostname] = Host{
			Hostname:                 hostname,
			SessionAverageLatency:    0,
			FiveMinuteAverageLatency: 0,
			FiveMinuteFailures:       0,
		}
		logs[hostname] = make([]CheckLog, 0)
		go checkHost(hostname, *portPtr, interval, channel)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
		}
	})

	go serve()

	for message := range channel {
		updateLog(message, &logs, &data.Hosts)
	}
}

func serve() {
	fmt.Println("Server is running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkHost(hostname string, port int, interval time.Duration, channel chan<- CheckLogMessage) {
	portStr := strconv.Itoa(port)
	for {
		startTime := time.Now()
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostname, portStr), time.Second)
		if err != nil {
			fmt.Println("Connection error:", hostname, err)
			channel <- CheckLogMessage{Hostname: hostname, Log: CheckLog{Timestamp: time.Now(), Latency: 0}}
		}
		if conn != nil {
			endTime := time.Now()
			var latency time.Duration = endTime.Sub(startTime)
			// fmt.Println("Opened", net.JoinHostPort(hostname, portStr))
			channel <- CheckLogMessage{Hostname: hostname, Log: CheckLog{Timestamp: time.Now(), Latency: uint16(latency.Milliseconds())}}
			conn.Close()
		} else {
			channel <- CheckLogMessage{Hostname: hostname, Log: CheckLog{Timestamp: time.Now(), Latency: 0}}
		}
		time.Sleep(interval)
	}
}

func updateLog(message CheckLogMessage, logs *map[string][]CheckLog, data *map[string]Host) {
	(*logs)[message.Hostname] = append((*logs)[message.Hostname], message.Log)
	host := (*data)[message.Hostname]
	if message.Log.Latency > 0 {
		host.LastSeen = message.Log.Timestamp
	}

	var fiveMinuteCalls uint = 0
	var sessionLatency []uint16 = make([]uint16, 0)
	var fiveMinuteLatency []uint16 = make([]uint16, 0)
	cutoff := time.Now().Add(-5 * time.Minute)
	for _, l := range (*logs)[message.Hostname] {
		fiveMinutes := l.Timestamp.After(cutoff)
		if fiveMinutes {
			fiveMinuteCalls++
			if l.Latency > 0 {
				fiveMinuteLatency = append(fiveMinuteLatency, l.Latency)
			}
		} else if l.Latency > 0 {
			sessionLatency = append(sessionLatency, l.Latency)
		}
	}
	host.FiveMinuteAverageLatency = avgLatency(fiveMinuteLatency)
	host.SessionAverageLatency = avgLatency(sessionLatency)
	host.FiveMinuteFailures = uint64(fiveMinuteCalls - uint(len(fiveMinuteLatency)))
	fmt.Println(host)

	(*data)[message.Hostname] = host
}

func avgLatency(latencies []uint16) uint16 {
	if len(latencies) == 0 {
		return 0
	}
	var total uint64
	for _, l := range latencies {
		total = total + uint64(l)
	}
	return uint16(total / uint64(len(latencies)))
}
