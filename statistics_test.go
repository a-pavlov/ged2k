package main

import (
	"testing"
	"time"
)

func Test_Statistics(t *testing.T) {
	stat := NewStatistics()
	for i := 0; i < 50; i++ {
		stat.SendBytes(1000)
		stat.ReceiveBytes(2000)
		d, err := time.ParseDuration("1s")
		if err != nil {
			t.Errorf("Duration error: %v", err)
		}
		stat.SecondTick(d)
	}

	if stat.DownloadRate() != 1996 {
		t.Errorf("Download rate is not correct: %v", stat.DownloadRate())
	}

	if stat.UploadRate() != 996 {
		t.Errorf("Upload rate is not correct: %v", stat.UploadRate())
	}
}

func Test_StatisticsLongInterval(t *testing.T) {
	stat := NewStatistics()
	for i := 0; i < 50; i++ {
		stat.SendBytes(1000)
		stat.ReceiveBytes(2000)
		d, err := time.ParseDuration("2s")
		if err != nil {
			t.Errorf("Duration error: %v", err)
		}
		stat.SecondTick(d)
	}

	if stat.DownloadRate() != 996 {
		t.Errorf("Download rate is not correct: %v", stat.DownloadRate())
	}

	if stat.UploadRate() != 496 {
		t.Errorf("Upload rate is not correct: %v", stat.UploadRate())
	}
}
