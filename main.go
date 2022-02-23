package main

import "fmt"

func main() {
	fmt.Println("【使用说明】")
	fmt.Println("1.开启监控 http://localhost:7777/ping/{host}  例如: http://localhost:7777/ping/www.tiktok.com")
	fmt.Println("2.关闭监控并展示结果 http://localhost:7777/pong/{host}  例如: http://localhost:7777/pong/www.tiktok.com")
	fmt.Println("3.关闭所有监控保存结果 http://localhost:7777/stop")
	fmt.Println("@监控结果保存在 report 文件夹里")

	httpServer()
}
