package main

import (
	"github.com/a-pavlov/ged2k/proto"
	"testing"
	"time"
)

func Test_Peers(t *testing.T) {
	e1, _ := proto.FromString("192.168.1.1:3333")
	e2, _ := proto.FromString("212.168.1.1:3330")
	e3, _ := proto.FromString("10.0.0.11:3330")
	p1 := Peer{endpoint: e1}
	p2 := Peer{endpoint: e2}
	p3 := Peer{endpoint: e3}

	if !LeftBetterRightToConnect(&p1, &p2) {
		t.Errorf("p1 not better than p2 - wrong")
	}

	if LeftBetterRightToConnect(&p3, &p1) {
		t.Errorf("p3 better than p1 - wrong")
	}

}

func Test_PolicyBase(t *testing.T) {
	policy := NewPolicy(4)
	p1 := policy.FindConnectCandidate(time.Now())
	if p1 != nil {
		t.Errorf("Connect candidate?")
	}

	e1, _ := proto.FromString("192.168.1.1:3333")
	e2, _ := proto.FromString("212.168.1.1:3330")
	e3, _ := proto.FromString("10.0.0.11:3330")
	e4, _ := proto.FromString("217.221.1.1:3330")

	if !policy.AddPeer(&Peer{endpoint: e1}) {
		t.Errorf("Can not add peer")
	}

	if !policy.AddPeer(&Peer{endpoint: e2}) {
		t.Errorf("Can not add peer")
	}

	if !policy.AddPeer(&Peer{endpoint: e3}) {
		t.Errorf("Can not add peer")
	}

	if !policy.AddPeer(&Peer{endpoint: e4}) {
		t.Errorf("Can not add peer")
	}

	if policy.AddPeer(&Peer{endpoint: e3}) {
		t.Errorf("Peer was added as duplicate")
	}

	tm := time.Now()
	peer1 := policy.FindConnectCandidate(tm)
	if peer1 == nil {
		t.Errorf("Can not find connect candidate")
	}

	peer1.peerConnection = &PeerConnection{}

	peer2 := policy.FindConnectCandidate(tm)
	if peer2 == nil {
		t.Errorf("Can not find connect candidate 2")
	}

	if peer1 == peer2 {
		t.Errorf("Connect candidates equal")
	}

	if !peer1.endpoint.IsLocalAddress() || !peer2.endpoint.IsLocalAddress() {
		t.Errorf("Wrong connect candidates")
	}

	peer2.NextConnection = tm.Add(time.Second * time.Duration(2))

	peer3 := policy.FindConnectCandidate(tm)
	peer3.peerConnection = &PeerConnection{}

	peer4 := policy.FindConnectCandidate(tm)
	peer4.peerConnection = &PeerConnection{}

	peer5 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(1)))

	if peer5 != nil {
		t.Errorf("Found connectot candidate when candidates exchosted")
	}

	peer6 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(3)))
	if peer6 != peer2 {
		t.Errorf("Reusing peer after next connection doesn't work")
	}

	peer6.NextConnection = tm.Add(time.Second * time.Duration(6))

	peer1.peerConnection = nil
	peer1.LastConnected = tm.Add(time.Second * time.Duration(4))
	peer1.FailCount += 1

	if policy.FindConnectCandidate(tm.Add(time.Second*time.Duration(5))) != nil {
		t.Errorf("Found unexpected connect candidate")
	}

	peer8 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(7)))

	if peer6 != peer8 {
		t.Errorf("Can not take peer after next connect")
	}

	peer8.peerConnection = &PeerConnection{}

	peer9 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(MIN_RECONNECT_TIMEOUT_SEC+5)))
	if peer9 != peer1 {
		t.Errorf("Can not take peer after fail")
	}

	peer9.peerConnection = &PeerConnection{}

	peer3.peerConnection = nil
	peer3.FailCount = 3

	peer4.peerConnection = nil
	peer4.FailCount = 2

	peer10 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(MIN_RECONNECT_TIMEOUT_SEC*6)))
	if peer10 != peer4 {
		t.Errorf("Wrong peer by fail count")
	}

	peer10.peerConnection = &PeerConnection{}

	peer11 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(MIN_RECONNECT_TIMEOUT_SEC*6)))
	if peer11 != peer3 {
		t.Errorf("Wrong peer by fail count 2")
	}

	e5, _ := proto.FromString("217.120.1.1:3330")
	e6, _ := proto.FromString("217.120.1.2:3330")

	if policy.AddPeer(&Peer{endpoint: e5}) {
		t.Errorf("Peer was added - error")
	}

	peer3.peerConnection = nil
	peer4.peerConnection = nil

	if policy.AddPeer(&Peer{endpoint: e5}) {
		t.Errorf("Peer was added - error")
	}

	peer3.FailCount = 6
	peer4.FailCount = 7

	if !policy.AddPeer(&Peer{endpoint: e5}) {
		t.Errorf("Peer was added - error")
	}

	if !policy.AddPeer(&Peer{endpoint: e6}) {
		t.Errorf("Peer was added - error")
	}

	peerA1 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(MIN_RECONNECT_TIMEOUT_SEC*7)))
	if peerA1.endpoint != e5 && peerA1.endpoint != e6 {
		t.Errorf("New added peer not found")
	}

	peerA1.peerConnection = &PeerConnection{}

	peerA2 := policy.FindConnectCandidate(tm.Add(time.Second * time.Duration(MIN_RECONNECT_TIMEOUT_SEC*7)))
	if peerA2.endpoint != e5 && peerA2.endpoint != e6 {
		t.Errorf("New added peer not found 2")
	}

	if policy.erasePeers() {
		t.Errorf("Erase peer returned true - error")
	}

}
