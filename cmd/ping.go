package main

import (
	"flag"
	"fmt"
	"github.com/moyada/smoke-ping/v2/log"
	"github.com/moyada/smoke-ping/v2/monitor"
	"github.com/moyada/smoke-ping/v2/util"
	"os"
)

func main() {
	var (
		host   = flag.String("host", "", "Address on which to monitor latency metrics.")
		size   = flag.Int("size", 1024, "Size of packet being sent.")
		output = flag.String("output", "", "Output location of the latency report.")
	)

	if len(os.Args) < 2 {
		fmt.Println("host require!!")
		return
	}

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

	if !util.IsValidIpAddress(addr) {
		fmt.Printf("host %v is invalid!!", addr)
		return
	}

	task := monitor.Task{Host: addr, Size: *size, Output: *output, Logger: &log.Console{}, Collector: &monitor.Chart{}}
	err := task.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
}
