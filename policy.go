package main

import (
	"log"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

const MAX_PEER_LIST_SIZE int = 100
const MIN_RECONNECT_TIMEOUT_SEC = 10

//const MAX_ITERATIONS = 50

const PEER_SRC_INCOMING byte = 0x1
const PEER_SRC_SERVER byte = 0x2
const PEER_SRC_DHT byte = 0x4
const PEER_SRC_RESUME_DATA byte = 0x8

type Peer struct {
	SourceFlag     byte
	LastConnected  time.Time
	NextConnection time.Time
	FailCount      int
	peerConnection *PeerConnection
	endpoint       proto.Endpoint
	Speed          int
}

func (p Peer) IsEmpty() bool {
	e := proto.Endpoint{}
	return p.endpoint == e
}

func (p *Peer) IsConnectCandidate() bool {
	return !(p.peerConnection != nil || p.FailCount > 5)
}

func (p *Peer) IsEraseCandidate() bool {
	if p.peerConnection != nil || p.IsConnectCandidate() {
		return false
	}

	return p.FailCount > 0
}

func (p *Peer) ShouldEraseImmediately() bool {
	return (p.SourceFlag & PEER_SRC_RESUME_DATA) == PEER_SRC_RESUME_DATA
}

func (p *Peer) SourceRank() int {
	ret := 0
	if (p.SourceFlag & PEER_SRC_SERVER) == PEER_SRC_SERVER {
		ret |= 1 << 5
	}

	if (p.SourceFlag & PEER_SRC_SERVER) == PEER_SRC_DHT {
		ret |= 1 << 4
	}

	if (p.SourceFlag & PEER_SRC_INCOMING) == PEER_SRC_INCOMING {
		ret |= 1 << 3
	}

	if (p.SourceFlag & PEER_SRC_RESUME_DATA) == PEER_SRC_RESUME_DATA {
		ret |= 1 << 2
	}

	return ret
}

func LeftBetterRightToRemove(l *Peer, r *Peer) bool {
	if l.FailCount != r.FailCount {
		return l.FailCount > r.FailCount
	}

	lResumeDataSource := (l.SourceFlag & PEER_SRC_RESUME_DATA) == PEER_SRC_RESUME_DATA
	rResumeDataSource := (r.SourceFlag & PEER_SRC_RESUME_DATA) == PEER_SRC_RESUME_DATA

	// prefer to drop peers whose only source is resume data
	if lResumeDataSource != rResumeDataSource {
		return lResumeDataSource
	}

	return false
}

type Policy struct {
	peers    map[proto.Endpoint]*Peer
	maxPeers int
}

func NewPolicy(mp int) *Policy {
	return &Policy{peers: make(map[proto.Endpoint]*Peer), maxPeers: mp}
}

func MakePolicy(mp int) Policy {
	return Policy{peers: make(map[proto.Endpoint]*Peer), maxPeers: mp}
}

func (policy *Policy) AddPeer(p *Peer) bool {
	if len(policy.peers) >= policy.maxPeers {
		if !policy.erasePeers() {
			return false
		}
	}

	oldPeer, ok := policy.peers[p.endpoint]
	if ok {
		oldPeer.SourceFlag |= p.SourceFlag
		return false
	}

	policy.peers[p.endpoint] = p
	return true
}

func (policy *Policy) erasePeers() bool {
	count := len(policy.peers)

	if count == 0 {
		return false
	}

	lowWatermark := policy.maxPeers * 95 / 100
	if lowWatermark == policy.maxPeers {
		lowWatermark--
	}

	eraseCandidate := proto.Endpoint{}

	for endpoint, peer := range policy.peers {
		if len(policy.peers) < lowWatermark {
			break
		}

		if peer.IsEraseCandidate() && (eraseCandidate.IsEmpty() || !LeftBetterRightToRemove(policy.peers[eraseCandidate], peer)) {
			eraseCandidate = endpoint
		}
	}

	if !eraseCandidate.IsEmpty() {
		delete(policy.peers, eraseCandidate)
	}

	return count != len(policy.peers)
}

func (policy *Policy) newConnection(connection *PeerConnection) bool {
	peer, ok := policy.peers[connection.endpoint]
	if ok {
		if peer.peerConnection != nil {
			log.Printf("peer %s already has peer connection\n", peer.endpoint.AsString())
			return false
		}

		peer.peerConnection = connection
		return true
	}

	p := Peer{endpoint: connection.endpoint, peerConnection: connection, SourceFlag: PEER_SRC_INCOMING}
	return policy.AddPeer(&p)
}

/**
 *
 * @param lhs
 * @param rhs
 * @return true if lhs better connect candidate than rhs
 */
func LeftBetterRightToConnect(l *Peer, r *Peer) bool {
	// prefer peers with lower failcount
	if l.FailCount != r.FailCount {
		return l.FailCount < r.FailCount
	}

	// Local peers should always be tried first
	lhsLocal := l.endpoint.IsLocalAddress()
	rhsLocal := r.endpoint.IsLocalAddress()
	if lhsLocal != rhsLocal {
		return lhsLocal
	}

	if l.LastConnected != r.LastConnected {
		return l.LastConnected.Before(r.LastConnected)
	}

	if l.NextConnection != r.NextConnection {
		return l.NextConnection.Before(r.NextConnection)
	}

	if l.SourceRank() != r.SourceRank() {
		return l.SourceRank() > r.SourceRank()
	}

	return false
}

func (policy *Policy) NumConnectCandidates() int {
	res := 0
	for _, x := range policy.peers {
		if x.IsConnectCandidate() {
			res++
		}
	}
	return res
}

func (policy *Policy) FindConnectCandidate(t time.Time) *Peer {
	candidate := proto.Endpoint{}

	for endpoint, peer := range policy.peers {
		if !peer.IsConnectCandidate() {
			continue
		}

		if !candidate.IsEmpty() && LeftBetterRightToConnect(policy.peers[candidate], peer) {
			continue
		}

		// 10 seconds timeout for each fail
		if !peer.LastConnected.IsZero() && t.Before(peer.LastConnected.Add(time.Second*time.Duration(peer.FailCount*MIN_RECONNECT_TIMEOUT_SEC))) {
			continue
		}

		if !peer.NextConnection.IsZero() && t.Before(peer.NextConnection) {
			continue
		}

		candidate = endpoint
	}

	return policy.peers[candidate]
}

func (policy *Policy) PeerConnectionClosed(peerConnection *PeerConnection, err error) {
	if peerConnection.peer != nil {
		p, ok := policy.peers[peerConnection.endpoint]
		if ok {
			p.LastConnected = time.Now()
			p.peerConnection = nil
			if err != nil {
				p.FailCount += 1
			}
		}
	}
}
