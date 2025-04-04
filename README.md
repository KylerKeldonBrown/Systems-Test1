This tool is a concurrent TCP port scanner written in Go.
It can scan a range of ports or specific ports across multiple targets simultaneously, and supports:
1.Concurrency using Goroutines and worker pools
2.Custom timeouts
3.Optional JSON output
4.Banner grabbing (tries to read any initial response from open ports)


Make sure you're in the same directory as main.go, then run:
go run -o portscanner main.go -This will generate a binary called portscanner.


Sample output
Scanning port 80 on scanme.nmap.org...
[OPEN] scanme.nmap.org:80
Scanning port 443 on scanme.nmap.org...
[CLOSED] scanme.nmap.org:443
=== Scan Summary ===
Targets Scanned: 1
Ports Scanned: 2
Open Ports: 1
Scan Duration: 1.00234567s


