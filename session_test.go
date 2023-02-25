package main

import (
	"github.com/a-pavlov/ged2k/proto"
	"testing"
)

func Test_HelloAnswer(t *testing.T) {
	cfg := Config{ListenPort: 30000, Name: "TestGed2k", MaxConnections: 100, ClientName: "test"}
	session := Session{configuration: cfg}
	ha := session.CreateHelloAnswer()
	data := make([]byte, proto.DataSize(ha))
	sb := proto.StateBuffer{Data: data}
	sb.Write(ha)
	if sb.Error() != nil {
		t.Errorf("Can not write Hello Answer %v", sb.Error())
	} else {
		sb2 := proto.StateBuffer{Data: data}
		var ha2 proto.HelloAnswer
		sb2.Read(&ha2)
		if sb2.Error() != nil {
			t.Errorf("Read hello answer error %v\n", sb2.Error())
		}
	}
}
