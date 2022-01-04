package monitor

import (
	"fmt"
	"github.com/moyada/smoke-ping/v2/ping"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var timeoutPacket = time.Duration(-1)

type Task struct {
	Host   string
	Size   int
	Output string

	startTime time.Time
	pinger    *ping.Pinger

	records map[int]*time.Duration
	report  Report

	recording int
	Collector Collector
}

type Report struct {
	MinRtt *time.Duration
	MaxRtt *time.Duration
	AvgRtt *time.Duration
	Loss   float64
}

func initTask(task *Task) error {
	if task.Size < 32 {
		task.Size = 32
	}

	task.recording = -1
	task.records = make(map[int]*time.Duration)
	task.report = Report{}

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

func (task *Task) Start() error {
	err := initTask(task)
	if err != nil {
		return err
	}

	done := make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("\nlatency monitor %s \n", sig)
			task.done()
			done <- true
		}
	}()

	task.startTime = time.Now()
	_, err = task.run()

	// wait for interrupt
	<-done

	return err
}

func (task *Task) run() (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(task.Host)
	if err != nil {
		return nil, err
	}
	//pinger.SetPrivileged(true)
	pinger.Size = task.Size - 8

	pinger.OnSend = func(pkt *ping.Packet) {
		task.records[pkt.Seq] = nil
	}

	pinger.OnTimeout = func(seq int) {
		fmt.Printf("Request timeout for icmp_seq %v\n", seq)
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
		task.records[pkt.Seq] = &pkt.Rtt
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

func (task *Task) done() {
	task.pinger.Stop()
	task.gather()
	task.saveResult()
}

func (task *Task) gather() {
	stats := task.pinger.Statistics()

	if task.report.AvgRtt == nil {
		task.report.AvgRtt = &stats.AvgRtt
		task.report.Loss = stats.PacketLoss
	} else {
		avg := (*task.report.AvgRtt + stats.AvgRtt) / 2
		task.report.AvgRtt = &avg
		task.report.Loss = (task.report.Loss + stats.PacketLoss) / 2
	}

	if task.report.MinRtt == nil || *task.report.MinRtt > stats.MinRtt {
		task.report.MinRtt = &stats.MinRtt
	}

	if task.report.MaxRtt == nil || *task.report.MaxRtt < stats.MaxRtt {
		task.report.MaxRtt = &stats.MaxRtt
	}
}

func toFileName(host string, startTime *time.Time, duration int) string {
	s1 := startTime.Format("2006-01-02 15:04:05")
	s2 := startTime.Add(time.Duration(duration) * time.Second).Format("15:04:05")
	return fmt.Sprintf("%v %v~%v.png", host, s1, s2)
}

func (task *Task) getOutput() string {
	if task.Output == "" {
		dir, _ := os.Getwd()
		return dir + "/" + toFileName(task.Host, &task.startTime, len(task.records))
	}
	dirIdx := strings.LastIndex(task.Output, "/")
	if dirIdx == -1 {
		dir, _ := os.Getwd()
		return dir + "/" + task.Output
	}

	output := task.Output
	path := task.Output[:dirIdx]
	if path[0] != '/' {
		dir, _ := os.Getwd()
		path = dir + "/" + path
		output = dir + "/" + output
	}

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return output
}

func (task *Task) saveResult() {
	count := len(task.records)
	if count < 2 {
		return
	}

	output := task.getOutput()
	fmt.Printf("build %v latency report >>> %v\n", task.Host, output)
	err := task.Collector.output(output, &task.startTime, task.records, &task.report)
	if err != nil {
		panic(err)
	}
}
