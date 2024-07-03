package main

import (
	"fmt"
	"runtime"
	"time"
)

type Test struct {
	ch   chan int
	quit chan struct{}
}

func (t *Test) bar() {
	timer := time.NewTimer(time.Second * 3)
	defer timer.Stop()
	for {
		select {
		case <-t.quit:
			fmt.Println("quit")
			return
		case <-timer.C: // 阻塞3s
		}

		fmt.Println("Hello")
		timer.Reset(time.Second * 3)
	}
}

func main() {
	quit := make(chan struct{}) // 用来判断是否退出的无缓冲 channel

	t := &Test{
		ch:   make(chan int),
		quit: make(chan struct{}),
	}

	go func() {
		time.Sleep(time.Second * 4)
		close(quit)
		runtime.Goexit()
	}()

	t.quit = quit
	t.bar()
}
