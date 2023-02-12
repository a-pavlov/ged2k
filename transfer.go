package main

import (
	"sync"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

type Transfer struct {
	mutex              sync.Mutex
	pause              bool
	stop               bool
	hashSet            []proto.Hash
	needSaveResumeData bool
	H                  proto.Hash
	connections        []*PeerConnection
	policy             Policy
	piecePicker        PiecePicker
	waitGroup          sync.WaitGroup
	commChan           chan string
	sourcesChan        chan proto.FoundFileSources
	stat               Statistics
}

func removePeerConnection(peerConnection *PeerConnection, pc []*PeerConnection) []*PeerConnection {
	var res []*PeerConnection
	for _, x := range pc {
		if x != peerConnection {
			res = append(res, x)
		}
	}

	return res
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

/*
func (t *Transfer) ConnectToPeer(peer Peer) {
	peer.LastConnected = time.Now()
	peer.NextConnection = time.Time{}
	connection := PeerConnection{session: t.session, peer: peer, transfer: t}
	connection.peer.peerConnection = &connection
	t.connections = append(t.connections, &connection)
	//session.connections.add(c)
	connection.Connect()
}

func (t *Transfer) PeerConnectionClose(peerConnection *PeerConnection, e error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.connections = removePeerConnection(peerConnection, t.connections)
	t.policy.PeerConnectionClosed(peerConnection, e)
}

func (t *Transfer) Tick(time time.Time, s *Session) {
	if !t.pause && !t.IsFinished() {
		for i := 0; i < 3; i++ {
			peer := t.policy.FindConnectCandidate(time)
			if !peer.IsEmpty() {
				peerConnection := s.ConnectoToPeer(peer.endpoint)
				if peerConnection != nil {
					// set peer connection to peer in policy
					//t.policy.
				}
			}
		}
	}
}*/

func (t *Transfer) WantMorePeers() bool {
	return !t.pause && !t.IsFinished() && t.policy.NumConnectCandidates() > 0
}

/*
func (t *Transfer) Tick() {
	tick := time.Tick(5000 * time.Millisecond)
	execute := true
	t.waitGroup.Add(1)
	defer t.waitGroup.Done()
	peerConnectionClosedChan := make(chan *PeerConnection)

E:
	for execute {
		select {
		case peerConn := <-peerConnectionClosedChan:
			t.policy.PeerConnectionClosed(peerConn, peerConn.lastError)
			t.session.ClosePeerConnection(peerConn.endpoint)
		case _, ok := <-t.sourcesChan:
			if ok {
				// pass sources to the policy
			}
		case comm, ok := <-t.commChan:
			if !ok {
				fmt.Println("Transfer exit requested")
				break E
			}

			switch comm {
			case "stop":
			// stop processing
			case "pause":
				t.pause = true
			case "resume":
				t.pause = false
			}
		case <-tick:
			fmt.Println("Transfer tick")
			if !t.pause && !t.IsFinsihed() {
				for i := 0; i < 3; i++ {
					peer := t.policy.FindConnectCandidate(time.Now())
					if !peer.IsEmpty() {
						peerConnection := t.session.ConnectoToPeer(peer.endpoint)
						if peerConnection != nil {
							// set peer connection to peer in policy
							//t.policy.
						}
					}
				}
			}
		}
	}
}*/

func (t *Transfer) IsFinished() bool {
	return t.piecePicker.NumHave() == t.piecePicker.PiecesCount()
}

func (t *Transfer) SecondTick(duration time.Duration, s *Session) {
	for _, x := range t.connections {
		t.stat.Add(&x.stat)
	}

	s.stat.Add(&t.stat)
	t.stat.SecondTick(duration)
}
