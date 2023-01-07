package main

import (
	"fmt"
	"net"

	"github.com/a-pavlov/ged2k/proto"
)

type ServerConnection struct {
	buffer         []byte
	connection     net.Conn
	packet_channel chan interface{}
}

func (sc *ServerConnection) Start() {
	connection, err := net.Dial("tcp", "5.45.85.226:6584")
	if err != nil {
		fmt.Println("Connect error", err)
		return
	}

	fmt.Println("Connected!")
	sc.connection = connection

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

	stateBuffer := proto.StateBuffer{Data: sc.buffer[proto.HEADER_SIZE:]}
	stateBuffer.Write(hello)
	if stateBuffer.Error() != nil {
		fmt.Printf("Error on serialize hello %v\n", stateBuffer.Error())
	}

	ph := proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: uint32(stateBuffer.Offset() + 1), Packet: proto.OP_LOGINREQUEST}
	ph.Write(sc.buffer)

	fmt.Printf("PACKET %x\n", sc.buffer[:stateBuffer.Offset()+proto.HEADER_SIZE])
	n, err := connection.Write(sc.buffer[:stateBuffer.Offset()+proto.HEADER_SIZE])
	fmt.Printf("Bytes %d have been written\n", n)
	if err != nil {
		fmt.Printf("Error write to socket %v\n", err)
	} else {

		pc := proto.PacketCombiner{}
		for {
			ph, bytes, error := pc.Read(connection)
			if error != nil {
				fmt.Printf("Can not read bytes from server %v", error)
			} else {
				fmt.Printf("Bytes from server %x count %d --> ", bytes, len(bytes))
			}

			sb := proto.StateBuffer{Data: bytes}

			switch ph.Packet {
			case proto.OP_SERVERLIST:
				elems, err := sb.ReadUint8()
				if err == nil && elems < 100 {
					c := proto.Collection{}
					for i := 0; i < int(elems); i++ {
						c = append(c, &proto.Endpoint{})
					}
					sb.Read(&c)
				}
			case proto.OP_GETSERVERLIST:
				// ignore
			case proto.OP_SERVERMESSAGE:
				bc := proto.ByteContainer{}
				bc.Get(&sb)
				if sb.Error() == nil {
					fmt.Println("Receive message from server", string(bc))
				}
			case proto.OP_SERVERSTATUS:
				ss := proto.Status{}
				ss.Get(&sb)
				if sb.Error() == nil {
					fmt.Println("Server status files:", ss.FilesCount, "users", ss.UsersCount)
				}
			case proto.OP_IDCHANGE:
				idc := proto.IdChange{}
				idc.Get(&sb)
				if sb.Error() == nil {
					fmt.Println("Server id change", idc.ClientId)
				}
			case proto.OP_SERVERIDENT:
				p := proto.UsualPacket{}
				p.Get(&sb)
				if sb.Error() == nil {
					fmt.Println("Received server info packet")
				}
			case proto.OP_SEARCHRESULT:
				p := proto.SearchResult{}
				p.Get(&sb)
				if sb.Error() == nil {
					fmt.Printf("Search result received: %d, more results %v\n", len(p.Items), p.MoreResults)
				} else {
					fmt.Printf("Unable to de-serealize %v", sb.Error())
				}
			case proto.OP_SEARCHREQUEST:
				// ignore
			case proto.OP_QUERY_MORE_RESULT:
				// ignore - out only
			case proto.OP_GETSOURCES:
				// ignore - out only
			case proto.OP_FOUNDSOURCES:
				// ignore
			case proto.OP_CALLBACKREQUEST:
				// ignore - out
			case proto.OP_CALLBACKREQUESTED:
				// ignore
			case proto.OP_CALLBACK_FAIL:
				// ignore
			default:
				fmt.Printf("Packet %x", bytes)
			}

			if sb.Error() != nil {
				fmt.Printf("Error on packet read %v", sb.Error())
			}

			/*
				switch(bytes[0]) {
					.OP_SERVERLIST.value, ServerList.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_GETSERVERLIST.value, GetList.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_SERVERMESSAGE.value, Message.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_SERVERSTATUS.value, Status.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_IDCHANGE.value, IdChange.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_SERVERIDENT.value, ServerInfo.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_SEARCHRESULT.value, SearchResult.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_SEARCHREQUEST.value, SearchRequest.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_QUERY_MORE_RESULT.value, SearchMore.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_GETSOURCES.value, GetFileSources.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_FOUNDSOURCES.value, FoundFileSources.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_CALLBACKREQUEST.value, CallbackRequest.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_CALLBACKREQUESTED.value, CallbackRequestIncoming.class);
					addHandler(ProtocolType.OP_EDONKEYHEADER.value, ClientServerTcp.OP_CALLBACK_FAIL.value, CallbackRequestFailed.class);
				}
			}*/
		}
	}
}

func (sc *ServerConnection) Search(s string) {
	parsed, err := proto.BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, s)
	for i := 0; i < 2; i++ {
		if err == nil {
			req, err := proto.PackRequest(parsed)
			if err == nil {
				stateBuffer := proto.StateBuffer{Data: sc.buffer[proto.HEADER_SIZE:]}
				for _, s := range req {
					s.Put(&stateBuffer)
				}

				if stateBuffer.Error() != nil {
					fmt.Printf("Error on serialize search %v\n", stateBuffer.Error())
				} else {
					ph := proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: uint32(stateBuffer.Offset() + 1), Packet: proto.OP_SEARCHREQUEST}
					ph.Write(sc.buffer)
					n, err := sc.connection.Write(sc.buffer[:stateBuffer.Offset()+proto.HEADER_SIZE])
					if err != nil {
						fmt.Printf("Error on send request %v\n", err)
					} else {
						fmt.Printf("Bytes %d have been written\n", n)
					}
				}
			}
		}
	}

}
