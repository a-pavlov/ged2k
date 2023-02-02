package main

import (
	"sync"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

type Transfer struct {
	mutex              sync.Mutex
	pause              bool
	session            *Session
	hashSet            []proto.Hash
	needSaveResumeData bool
	H                  proto.Hash
	connections        []*PeerConnection
	policy             Policy
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

func (t *Transfer) AttachPeer(connection *PeerConnection) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.policy.newConnection(connection)
	t.connections = append(t.connections, connection)
	connection.transfer = t
}

func (t *Transfer) ConnectToPeer(peer Peer) {
	peer.LastConnected = time.Now()
	peer.NextConnection = time.Time{}
	connection := PeerConnection{session: t.session, peer: peer, transfer: t}
	connection.peer.peerConnection = &connection
	t.connections = append(t.connections, &connection)
	//session.connections.add(c)
	connection.Connect()
}
