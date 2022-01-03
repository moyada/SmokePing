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
		host   	= flag.String("host", "", "Address on which to monitor delay metrics.")
		size   	= flag.Int("size", 1024, "Size of packet being sent.")
		output 	= flag.String("output", "", "Output location of the latency report.")
	)


	addr := os.Args[1]
	if addr == "" {
		fmt.Println("host require!!")
		return
	}

	if addr[0] == '-' {
		flag.Parse()
		addr = *host
	} else {
		flag.CommandLine.Parse(os.Args[2:])
	}

	if !isValidIpAddress(addr) {
		fmt.Printf("host %v is invalid!!", addr)
		return
	}

	task := monitor.Task{Host: addr, Size: *size, Output: *output, Collector: &monitor.Chart{}}
	err := task.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
}
