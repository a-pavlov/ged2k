package main

import "C"
import "sync"

type StatChannel struct {
	secondCounter int
	totalCounter  int
	average5Sec   int
	average30Sec  int
	samples       []int
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

func (sc *StatChannel) calc(timeIntervalMS int) {
	sample := (sc.secondCounter * 1000) / timeIntervalMS
	sc.samples = append(sc.samples, sample)
	if len(sc.samples) > 5 {
		sc.samples = sc.samples[1:]
	}
	sum := 0
	for _, x := range sc.samples {
		sum += x
	}

	sc.average5Sec = sum / 5
	//m_5_sec_average = size_type(m_5_sec_average) * 4 / 5 + sample / 5;
	sc.average30Sec = sc.average30Sec*29/30 + sample/30
	sc.secondCounter = 0
}

const CH_UPLOAD_PAYLOAD int = 0
const CH_UPLOAD_PROTOCOL int = 1
const CH_DOWNLOAD_PAYLOAD int = 2
const CH_DOWNLOAD_PROTOCOL int = 3

type Statistics struct {
	mutex    sync.Mutex
	channels []StatChannel
}

func (s *Statistics) Add(stat *Statistics) {
	s.mutex.Lock()
	s.mutex.Unlock()
	for i := range stat.channels {
		s.channels[i].AddChannel(stat.channels[i])
	}
}

func (s *Statistics) ReceiveBytes(protocolBytes int, payloadBytes int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.channels[CH_DOWNLOAD_PROTOCOL].Add(protocolBytes)
	s.channels[CH_DOWNLOAD_PAYLOAD].Add(payloadBytes)
}

func (s *Statistics) SendBytes(protocolBytes int, payloadBytes int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.channels[CH_UPLOAD_PROTOCOL].Add(protocolBytes)
	s.channels[CH_UPLOAD_PAYLOAD].Add(payloadBytes)
}

func (s *Statistics) DownloadRate() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.channels[CH_DOWNLOAD_PAYLOAD].Rate() + s.channels[CH_DOWNLOAD_PROTOCOL].Rate()
}

func (s *Statistics) UploadRate() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.channels[CH_UPLOAD_PAYLOAD].Rate() + s.channels[CH_UPLOAD_PROTOCOL].Rate()
}
