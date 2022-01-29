package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/moyada/smoke-ping/v2/monitor"
	"github.com/moyada/smoke-ping/v2/util"
	"io/ioutil"
	"net/http"
)

var hostJob = make(map[string]*monitor.WebTask)

func httpServer() {
	gin.ForceConsoleColor()
	engine := gin.Default()
	engine.GET("ping/:host", ping)
	engine.GET("pong/:host", pong)
	engine.GET("stop", stop)
	engine.Run(":7777")
}

func ping(c *gin.Context) {
	host := c.Param("host")
	if !util.IsValidIpAddress(host) {
		c.String(http.StatusBadRequest, "host invalid")
		return
	}

	task := monitor.Task{Host: host, Size: 64, Output: "report", Collector: &monitor.Chart{}}
	job := &monitor.WebTask{Task: task}

	hostJob[host] = job

	go job.Start()

	msg := fmt.Sprintf("start monitoring %v", task.Host)
	fmt.Println(msg)
	c.String(http.StatusOK, msg)
}

func pong(c *gin.Context) {
	host := c.Param("host")
	if !util.IsValidIpAddress(host) {
		c.String(http.StatusBadRequest, "host invalid")
		return
	}

	job, exist := hostJob[host]
	if !exist {
		c.String(http.StatusNotFound, "%v monitor not found", host)
		return
	}

	output, err := job.Done()
	if err != nil {
		c.Error(err)
		return
	}
	file, _ := ioutil.ReadFile(output)
	c.Writer.WriteString(string(file))
}

func stop(c *gin.Context) {
	for _, task := range hostJob {
		task.Done()
	}
	c.String(http.StatusOK, "ok")
}
