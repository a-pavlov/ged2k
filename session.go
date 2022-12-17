package main

import (
	"fmt"
	"sync"
	"time"
)

type Session struct {
	comm chan string
	wg   sync.WaitGroup
}

func (s *Session) Tick() {
	tick := time.Tick(5000 * time.Millisecond)
	execute := true
	s.wg.Add(1)
	defer s.wg.Done()

	for execute {
		select {
		case c := <-s.comm:
			switch c {
			case "quit":
				execute = false
			}
		case <-tick:
			fmt.Println("Tick")
		}
	}

	fmt.Println("Session closed")
}

func (s *Session) Wait() {
	s.comm <- "quit"
	s.wg.Wait()
}
