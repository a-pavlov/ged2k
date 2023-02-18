package main

import (
	"math/rand"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

const MAX_PEER_LIST_SIZE int = 100
const MIN_RECONNECT_TIMEOUT_SEC = 10
const MAX_ITERATIONS = 50

const PEER_SRC_INCOMING byte = 0x1
const PEER_SRC_SERVER byte = 0x2
const PEER_SRC_DHT byte = 0x4
const PEER_SRC_RESUME_DATA byte = 0x8

const PEER_SPEED_SLOW int = 0
const PEER_SPEED_MEDIUM int = 1
const PEER_SPEED_FAST int = 2

type Peer struct {
	SourceFlag     byte
	LastConnected  time.Time
	NextConnection time.Time
	FailCount      int
	Connectable    bool
	peerConnection *PeerConnection
	endpoint       proto.Endpoint
	Speed          int
}

func (p Peer) IsEmpty() bool {
	e := proto.Endpoint{}
	return p.endpoint == e
}

func (p *Peer) IsConnectCandidate() bool {
	return !(p.peerConnection != nil || !p.Connectable || p.FailCount > 10)
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

// true if left better to erase than right
func ComparePeerErase(l *Peer, r *Peer) bool {
	if l.FailCount != r.FailCount {
		return l.FailCount > r.FailCount
	}

	lResumeDataSource := (l.SourceFlag & PEER_SRC_RESUME_DATA) == PEER_SRC_RESUME_DATA
	rResumeDataSource := (r.SourceFlag & PEER_SRC_RESUME_DATA) == PEER_SRC_RESUME_DATA

	// prefer to drop peers whose only source is resume data
	if lResumeDataSource != rResumeDataSource {
		return lResumeDataSource
	}

	if l.Connectable != r.Connectable {
		return !l.Connectable
	}

	return false
}

type Policy struct {
	roundRobin int
	peers      []*Peer
	transfer   *Transfer
}

func (policy *Policy) AddPeer(p *Peer) bool {
	if len(policy.peers) >= MAX_PEER_LIST_SIZE {
		if !policy.erasePeers() {
			return false
		}
	}

	indx := policy.GetPeerIndexByEndpoint(p.endpoint)
	if indx != -1 {
		policy.peers[indx].SourceFlag |= p.SourceFlag
		return false
	}

	policy.peers = append(policy.peers, p)
	return true
}

func (policy *Policy) GetPeerIndexByEndpoint(ep proto.Endpoint) int {
	for i, x := range policy.peers {
		if x.endpoint == ep {
			return i
		}
	}

	return -1
}

// costly
func removePeer(s []*Peer, i int) []*Peer {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (policy *Policy) erasePeers() bool {

	count := len(policy.peers)

	if count == 0 {
		return false
	}

	eraseCandidate := -1

	roundRobin := rand.Intn(len(policy.peers))

	lowWatermark := MAX_PEER_LIST_SIZE * 95 / 100
	if lowWatermark == MAX_PEER_LIST_SIZE {
		lowWatermark--
	}

	for iterations := proto.Min(len(policy.peers), MAX_ITERATIONS); iterations > 0; iterations-- {
		if len(policy.peers) < lowWatermark {
			break
		}

		if roundRobin == len(policy.peers) {
			roundRobin = 0
		}

		p := policy.peers[roundRobin]
		current := roundRobin

		// check p is erase candidate or we already have erase candidate and it not better than pe for erase
		if p.IsEraseCandidate() && (eraseCandidate == -1 || !ComparePeerErase(policy.peers[eraseCandidate], p)) {
			if p.ShouldEraseImmediately() {
				if eraseCandidate > current {
					eraseCandidate--
				}

				policy.peers = removePeer(policy.peers, current)
			} else {
				eraseCandidate = current
			}
		}

		roundRobin++
	}

	if eraseCandidate > -1 {
		policy.peers = removePeer(policy.peers, eraseCandidate)
	}

	return count != len(policy.peers)
}

func (policy *Policy) newConnection(connection *PeerConnection) bool {
	indx := policy.GetPeerIndexByEndpoint(connection.peer.endpoint)
	if indx != -1 {

		p := policy.peers[indx]
		if p.peerConnection != nil {
			return false
		}

		p.peerConnection = connection
		return true
	}

	p := Peer{endpoint: connection.peer.endpoint, peerConnection: connection}
	return policy.AddPeer(&p)
}

/**
 *
 * @param lhs
 * @param rhs
 * @return true if lhs better connect candidate than rhs
 */
func comparePeers(l *Peer, r *Peer) bool {
	// prefer peers with lower failcount
	if l.FailCount != r.FailCount {
		return l.FailCount < r.FailCount
	}

	// Local peers should always be tried first
	//boolean lhsLocal = Utils.isLocalAddress(lhs.getEndpoint());
	//boolean rhsLocal = Utils.isLocalAddress(rhs.getEndpoint());
	//if (lhsLocal != rhsLocal) {
	//    return lhsLocal;
	//}

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
	candidate := -1
	eraseCandidate := -1
	if policy.roundRobin >= len(policy.peers) {
		policy.roundRobin = 0
	}

	for iteration := 0; iteration < proto.Min(len(policy.peers), MAX_ITERATIONS); iteration++ {
		if policy.roundRobin >= len(policy.peers) {
			policy.roundRobin = 0
		}

		p := policy.peers[policy.roundRobin]
		current := policy.roundRobin

		if len(policy.peers) > MAX_PEER_LIST_SIZE {
			if p.IsEraseCandidate() && (eraseCandidate == -1 || !ComparePeerErase(policy.peers[eraseCandidate], p)) {
				if p.ShouldEraseImmediately() {
					if eraseCandidate > current {
						eraseCandidate--
					}

					if candidate > current {
						candidate--
					}

					policy.peers = removePeer(policy.peers, current)
					continue
				} else {
					eraseCandidate = current
				}
			}
		}

		policy.roundRobin++
		if !p.IsConnectCandidate() {
			continue
		}

		if candidate != -1 && comparePeers(policy.peers[candidate], p) {
			continue
		}

		if !p.NextConnection.IsZero() && p.NextConnection.Before(t) {
			continue
		}

		// 10 seconds timeout for each fail
		if !p.LastConnected.IsZero() && t.Before(p.LastConnected.Add(time.Second*time.Duration(p.FailCount*MIN_RECONNECT_TIMEOUT_SEC))) {
			continue
		}
		candidate = current
	}

	if eraseCandidate != -1 {
		if candidate > eraseCandidate {
			candidate--
		}

		policy.peers = removePeer(policy.peers, eraseCandidate)
	}

	if candidate == -1 {
		return nil
	}

	x := time.Now().Add(time.Second * time.Duration(5))
	policy.peers[candidate].NextConnection = x
	return policy.peers[candidate]
}

func (policy *Policy) PeerConnectionClosed(peerConnection *PeerConnection, e error) {
	for _, x := range policy.peers {
		if x.peerConnection == peerConnection {
			x.LastConnected = time.Now()
			if e != nil {
				x.FailCount = x.FailCount + 1
			}
			x.peerConnection = nil
			break
		}
	}
}

func (policy *Policy) NewConnection(peerConnection *PeerConnection) {
	for _, x := range policy.peers {
		if x.endpoint == peerConnection.endpoint {
			if x.peerConnection == nil {
				peerConnection.peer = x
				x.peerConnection = peerConnection
				return
			}
		}
	}

	peer := Peer{endpoint: peerConnection.endpoint, peerConnection: peerConnection}
	policy.AddPeer(&peer)
}
