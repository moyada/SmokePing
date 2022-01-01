package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"ping-prober/v2/monitor"
	"strings"
	"unicode"
)

func isValidIpAddress(addr string) bool {
	if addr == "" {
		return false
	}

	ip := net.ParseIP(addr)
	if ip != nil {
		return true
	}

	// web host
	pc := strings.Count(addr, ".")
	if pc < 1 || pc > 2 {
		return false
	}

	index := strings.Index(addr, ".")
	if index == 0 {
		return false
	}
	index = strings.LastIndex(addr, ".")
	end := addr[index+1:]
	if end == "" {
		return false
	}

	for _, t := range end {
		if !unicode.IsLetter(t) {
			return false
		}
	}
	return true
}

func main() {
	var (
		host   = flag.String("host", "", "Address on which to monitor delay metrics.")
		size   = flag.Int("size", 1024, "Size of packet being sent.")
		output = flag.String("output", "", "Output location of the report.")
	)
	flag.Parse()
	if *host == "" && len(os.Args) > 1 {
		host = &os.Args[1]
	}

	if !isValidIpAddress(*host) {
		err := fmt.Errorf("host %v is invalid!!", *host)
		fmt.Println(err)
		return
	}

	task := monitor.Task{Host: *host, Size: *size, Output: *output, Collector: &monitor.Chart{}}
	task.Start()
}
