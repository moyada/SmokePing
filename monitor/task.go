package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"ping-prober/v2/ping"
	"strings"
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

func initTask(task *Task) error {
	if task.Size < 32 {
		task.Size = 32
	}

	task.recording = -1

	records := make(map[int]*time.Duration, 0)
	task.records = &records

	task.report = &Report{}

	if task.Output != "" {
		prefixIdx := strings.LastIndex(task.Output, ".png")
		if prefixIdx < 1 || task.Output[prefixIdx:] != ".png" {
			return fmt.Errorf("output should be end with .png")
		}

		if task.Output[prefixIdx-1] == '/' {
			return fmt.Errorf("invalid output name")
		}
	}

	return nil
}

func toFileName(host string, startTime *time.Time, duration int) string {
	s1 := startTime.Format("2006-01-02 15:04:05")
	s2 := startTime.Add(time.Duration(duration) * time.Second).Format("15:04:05")
	return fmt.Sprintf("%v %v~%v.png", host, s1, s2)
}

func (task *Task) Start() error {
	err := initTask(task)
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("\nlatency monitor %s \n", sig)
			task.recordAll()
			task.saveResult()
			task.pinger.Stop()
			os.Exit(1)
		}
	}()

	t := time.Now()
	task.startTime = &t
	_, err = task.run(0)
	if err != nil {
		return err
	}

	//var index = 1
	//for {
	//	stat := task.run(index)
	//	task.addReport(stat)
	//
	//	go task.record(index)
	//	index = index + count
	//}
	return nil
}

func (task *Task) run(index int) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(task.Host)
	if err != nil {
		return nil, err
	}
	pinger.Size = task.Size - 8

	pinger.OnSend = func(pkt *ping.Packet) {
		(*task.records)[index+pkt.Seq] = nil
	}

	pinger.OnTimeout = func(seq int) {
		fmt.Printf("Request timeout for icmp_seq %v\n", seq)
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
		return nil, err
	}
	return pinger.Statistics(), nil
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
}

func (task *Task) saveResult() {
	count := len(*task.records)
	if count < 2 {
		return
	}

	if task.Output == "" {
		task.Output = toFileName(task.Host, task.startTime, len(*task.records))
	} else {
		dirIdx := strings.LastIndex(task.Output, "/")
		if dirIdx != -1 {
			os.MkdirAll(task.Output[:dirIdx], os.ModePerm)
		}
	}

	fmt.Printf("build %v latency report >>> %v\n", task.Host, task.Output)
	err := task.Collector.output(task.Output, task.startTime, task.records, task.report)
	if err != nil {
		panic(err)
	}
}
