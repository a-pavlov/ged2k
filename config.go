package main

import "github.com/a-pavlov/ged2k/proto"

type Config struct {
	ListenPort           uint16
	Name                 string
	UserAgent            proto.EMuleHash
	ClientName           string
	ModName              string
	AppVersion           uint32
	ModMajorVersion      uint32
	ModMinorVersion      uint32
	ModBuildVersion      uint32
	MaxConnectsPerSecond int
	MaxConnections       int
}
