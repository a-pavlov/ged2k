package main

import (
	"sync"

	"github.com/a-pavlov/ged2k/proto"
)

type Transfer struct {
	mutex              sync.Mutex
	pause              bool
	session            *Session
	hashSet            []proto.Hash
	needSaveResumeData bool
}

func (t *Transfer) IsPaused() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.pause
}

func (t *Transfer) IsNeedSaveResumeData() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.needSaveResumeData
}
