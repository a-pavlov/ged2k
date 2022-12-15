package main

import (
	"fmt"
	"net"

	"github.com/a-pavlov/ged2k/proto"
)

type ServerConnection struct {
	buffer []byte
}

func (sc *ServerConnection) Process() {
	fmt.Println("Connecting to", "5.45.85.226:6584")
	connection, err := net.Dial("tcp", "5.45.85.226:6584")
	//connection, err := net.Dial("tcp", "localhost:2000")
	if err != nil {
		fmt.Println("Connect error", err)
		return
	}

	fmt.Println("Connected!")

	var version uint32 = 0x3c
	var versionClient uint32 = (proto.GED2K_VERSION_MAJOR << 24) | (proto.GED2K_VERSION_MINOR << 17) | (proto.GED2K_VERSION_TINY << 10) | (1 << 7)
	var capability uint32 = proto.CAPABLE_AUXPORT | proto.CAPABLE_NEWTAGS | proto.CAPABLE_UNICODE | proto.CAPABLE_LARGEFILES | proto.CAPABLE_ZLIB

	fmt.Println("Version client", versionClient)

	var hello proto.UsualPacket
	hello.H = proto.EMULE
	hello.Point = proto.Endpoint{Ip: 0, Port: 20033}
	hello.Properties = append(hello.Properties, proto.CreateTag(version, proto.CT_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(capability, proto.CT_SERVER_FLAGS, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag("ged2k", proto.CT_NAME, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(versionClient, proto.CT_EMULE_VERSION, ""))

	stateBuffer := proto.StateBuffer{Data: sc.buffer[6:]}
	stateBuffer.Write(hello)
	if stateBuffer.Error() == nil {
		ph := proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: uint32(stateBuffer.Offset() + 1), Packet: proto.OP_LOGINREQUEST}
		sb2 := proto.StateBuffer{Data: sc.buffer}
		ph.Put(&sb2)
		if sb2.Error() == nil {
			fmt.Printf("PACKET %x\n", sc.buffer[:stateBuffer.Offset()+6])
			n, err := connection.Write(sc.buffer[:stateBuffer.Offset()+6])
			if err != nil {
				fmt.Printf("Error write to socket %v\n", err)
			} else {
				fmt.Printf("Wrote %d bytes as hello packet\n", n)
				pc := proto.PacketCombiner{}
				bytes, error := pc.Read(connection)
				if error != nil {
					fmt.Printf("Can not read bytes from server %v", error)
				} else {
					fmt.Printf("Bytes from server %x", bytes)
				}
			}
		} else {
			fmt.Printf("Serialize header error %v\n", sb2.Error())
		}
	} else {
		fmt.Printf("Error on serialize hello %v\n", stateBuffer.Error())
	}

	connection.Close()
}
