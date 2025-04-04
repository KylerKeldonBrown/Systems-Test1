// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go
// Kyler Brown Test1

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	Target string `json:"target"`
	Port   int    `json:"port"`
	Banner string `json:"banner,omitempty"`
	Open   bool   `json:"open"`
}

type ScanTask struct {
	Address string
	Target  string
	Port    int
}

func worker(wg *sync.WaitGroup, tasks chan ScanTask, results chan ScanResult, timeout time.Duration) {
	defer wg.Done()

	for task := range tasks {
		fmt.Printf("Scanning port %d on %s...\n", task.Port, task.Target)

		var conn net.Conn
		var err error

		for attempt := 1; attempt <= 2; attempt++ {
			dialer := net.Dialer{Timeout: timeout}
			conn, err = dialer.Dial("tcp", task.Address)
			if err == nil {
				break
			}
			backoff := time.Duration(1<<attempt) * time.Second
			fmt.Printf("Retry %d for %s after %v\n", attempt, task.Address, backoff)
			time.Sleep(backoff)
		}

		result := ScanResult{Target: task.Target, Port: task.Port, Open: err == nil}

		if err == nil {
			defer conn.Close()
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			buff := make([]byte, 1024)
			n, _ := conn.Read(buff)

			if n > 0 {
				result.Banner = string(buff[:n])
			} else {
				result.Banner = "No response"
			}
		}

		results <- result
	}
}

func parsePorts(ports string) []int {
	var portList []int
	split := strings.Split(ports, ",")
	for _, p := range split {
		portStr := strings.TrimSpace(p)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse port: %s\n", portStr)
			continue
		}
		portList = append(portList, port)
	}
	return portList
}

func main() {
	targetsFlag := flag.String("targets", "scanme.nmap.org", "Comma-separated list of target IPs/hostnames")
	startPort := flag.Int("start-port", 1, "Start of port range")
	endPort := flag.Int("end-port", 1024, "End of port range")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Int("timeout", 3, "Timeout in seconds for each connection")
	jsonOut := flag.Bool("json", false, "Output results in JSON format")
	portList := flag.String("ports", "", "Comma-separated list of specific ports to scan")
	flag.Parse()

	targets := strings.Split(*targetsFlag, ",")

	var ports []int
	if *portList != "" {
		ports = parsePorts(*portList)
	} else {
		for p := *startPort; p <= *endPort; p++ {
			ports = append(ports, p)
		}
	}

	var wg sync.WaitGroup
	tasks := make(chan ScanTask, 100)
	results := make(chan ScanResult, 1000)

	start := time.Now()

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, results, time.Duration(*timeout)*time.Second)
	}

	go func() {
		for _, target := range targets {
			trimmedTarget := strings.TrimSpace(target)
			for _, port := range ports {
				address := net.JoinHostPort(trimmedTarget, strconv.Itoa(port))
				tasks <- ScanTask{
					Address: address,
					Target:  trimmedTarget,
					Port:    port,
				}
			}
		}
		close(tasks)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var openPorts []ScanResult
	totalScanned := 0
	for res := range results {
		totalScanned++
		if res.Open {
			fmt.Printf("[OPEN] %s:%d\n", res.Target, res.Port)
			openPorts = append(openPorts, res)
		} else {
			fmt.Printf("[CLOSED] %s:%d\n", res.Target, res.Port)
		}
	}

	elapsed := time.Since(start)

	// Output
	fmt.Println("=== Scan Summary ===")
	fmt.Printf("Targets Scanned: %d\n", len(targets))
	fmt.Printf("Ports Scanned: %d\n", len(ports)*len(targets))
	fmt.Printf("Open Ports: %d\n", len(openPorts))
	fmt.Printf("Scan Duration: %v\n", elapsed)

	if *jsonOut {
		json.NewEncoder(os.Stdout).Encode(openPorts)
	}
}
