package monitor

import (
	"fmt"
	"os"
	"time"
)

type WebTask struct {
	Task
	output string
}

func initWebJob(task *WebTask) {
	if task.Size < 32 {
		task.Size = 32
	}

	task.recording = -1
	task.records = make(map[int]*time.Duration)
	task.report = Report{}
}

func (task *WebTask) Start() error {
	initWebJob(task)

	task.startTime = time.Now()
	_, err := task.run()

	return err
}

func (task *WebTask) Done() (string, error) {
	if task.output != "" {
		return task.output, nil
	}
	task.pinger.Stop()
	task.gather()
	output, err := task.saveResult()
	if err == nil {
		task.output = output
	}
	return output, err
}

func (task *WebTask) getOutput() string {
	dir, _ := os.Getwd()
	path := dir + "/" + task.Output

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return path + "/" + toFileName(task.Host, &task.startTime, len(task.records))
}

func (task *WebTask) saveResult() (string, error) {
	count := len(task.records)
	if count < 2 {
		return "", nil
	}

	output := task.getOutput()

	fmt.Printf("build %v latency report >>> %v\n", task.Host, output)

	err := task.Collector.output(output, &task.startTime, task.records, &task.report)
	return output, err
}
