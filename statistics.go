package main

import "C"
import (
	"time"
)

type StatChannel struct {
	secondCounter int
	totalCounter  int
	average5Sec   int
	average30Sec  int
}

func (sc *StatChannel) Add(count int) {
	sc.secondCounter += count
	sc.totalCounter += count
}

func (sc *StatChannel) Rate() int {
	return sc.average5Sec
}

func (sc *StatChannel) LowPassRate() int {
	return sc.average30Sec
}

func (sc *StatChannel) AddChannel(s StatChannel) *StatChannel {
	sc.secondCounter += s.secondCounter
	sc.totalCounter += s.totalCounter
	return sc
}

func (sc *StatChannel) calc(duration time.Duration) {
	sample := (sc.secondCounter * 1000) / int(duration.Milliseconds())
	sc.average5Sec = sc.average5Sec*4/5 + sample/5
	sc.average30Sec = sc.average30Sec*29/30 + sample/30
	sc.secondCounter = 0
}

const (
	CH_UPLOAD = iota
	CH_DOWNLOAD
)

type Statistics struct {
	channels []*StatChannel
}

func NewStatistics() *Statistics {
	return &Statistics{channels: []*StatChannel{{}, {}}}
}

func MakeStatistics() Statistics {
	return Statistics{channels: []*StatChannel{{}, {}}}
}

//func (s *Statistics) Add(stat Statistics) {
//for i := 0; i < len(s.channels); i++ {
//		s.channels[i].AddChannel(*stat.channels[i])
//}
//}

func (s *Statistics) SecondTick(duration time.Duration) {
	for _, x := range s.channels {
		x.calc(duration)
	}
}

func (s *Statistics) ReceiveBytes(bytes int) {
	s.channels[CH_DOWNLOAD].Add(bytes)
}

func (s *Statistics) SendBytes(bytes int) {
	s.channels[CH_UPLOAD].Add(bytes)
}

func (s *Statistics) DownloadRate() int {
	return s.channels[CH_DOWNLOAD].Rate()
}

func (s *Statistics) UploadRate() int {
	return s.channels[CH_UPLOAD].Rate()
}

func (s *Statistics) DownloadLowPassRate() int {
	return s.channels[CH_DOWNLOAD].LowPassRate()
}

func (s *Statistics) UploadLowPassRate() int {
	return s.channels[CH_UPLOAD].LowPassRate()
}
