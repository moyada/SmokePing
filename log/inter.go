package log

import "fmt"

type BizLogger interface {
	Info(msg string)
	Error(err error)
}

type Console struct {
}

func (c *Console) Info(msg string) {
	fmt.Println(msg)
}

func (c *Console) Error(err error) {
	fmt.Println(err)
}
