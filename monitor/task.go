package monitor

import (
	"fmt"
	"github.com/go-ping/ping"
	"os"
	"os/signal"
	"time"
)

type Task struct {
	Host   string
	Size   int
	Output string

	startTime *time.Time
	pinger    *ping.Pinger

	records *map[int]*time.Duration
	report  *Report

	recording int
	Collector Collector
}

type Report struct {
	MinRtt *time.Duration
	MaxRtt *time.Duration
	AvgRtt *time.Duration
}

const count = 60

func initTask(task *Task) {
	task.recording = -1

	records := make(map[int]*time.Duration, 0)
	task.records = &records

	task.report = &Report{}
}

func (task *Task) Start() {
	initTask(task)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("\nmonitor %s \n", sig)
			task.recordAll()
			os.Exit(1)
		}
	}()

	t := time.Now()
	task.startTime = &t
	task.run(1)

	//var index = 1
	//for {
	//	stat := task.run(index)
	//	task.addReport(stat)
	//
	//	go task.record(index)
	//	index = index + count
	//}
}

func (task *Task) run(index int) *ping.Statistics {
	pinger, err := ping.NewPinger(task.Host)
	if err != nil {
		panic(err)
	}
	pinger.Size = task.Size - 8

	pinger.OnSend = func(pkt *ping.Packet) {
		(*task.records)[index+pkt.Seq] = nil
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
		(*task.records)[index+pkt.Seq] = &pkt.Rtt
	}

	task.pinger = pinger
	//pinger.Count = count
	//pinger.Timeout = time.Second * count
	pinger.RecordRtts = false

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		panic(err)
	}
	return pinger.Statistics()
}

func (task *Task) record(index int) {
	end := index + count
	for i := index; i < end; i++ {
		t := task.startTime.Add(time.Second * time.Duration(i))
		timeout := (*task.records)[i]
		task.Collector.record(i, &t, timeout)

		//fmt.Printf("%v  %v\n", t, timeout)
		//task.recording = i
	}
}

func (task *Task) addReport(stat *ping.Statistics) {
	if task.report.AvgRtt == nil {
		task.report.AvgRtt = &stat.AvgRtt
	} else {
		avg := (*task.report.AvgRtt + stat.AvgRtt) / 2
		task.report.AvgRtt = &avg
	}

	if task.report.MinRtt == nil || *task.report.MinRtt > stat.MinRtt {
		task.report.MinRtt = &stat.MinRtt
	}

	if task.report.MaxRtt == nil || *task.report.MaxRtt < stat.MaxRtt {
		task.report.MaxRtt = &stat.MaxRtt
	}
}

func (task *Task) recordAll() {
	//var keys []int
	//for key := range *task.records {
	//	keys = append(keys, key)
	//}
	//
	//sort.Ints(keys)
	//for _, key := range keys {
	//	timeout := (*task.records)[key]
	//	t := task.startTime.Add(time.Second * time.Duration(key))
	//	fmt.Printf("%v  %v\n", t, timeout)
	//}

	stats := task.pinger.Statistics()
	task.addReport(stats)

	err := task.Collector.output(task.Host, task.startTime, task.records, task.report, task.Output)
	if err != nil {
		panic(err)
	}
}
