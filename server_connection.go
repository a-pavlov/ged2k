package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

const (
	Connected    = iota
	Connecting   = iota
	Disconnected = iota
)

type ServerConnection struct {
	mutex          sync.Mutex
	buffer         []byte
	connection     net.Conn
	status         int
	packet_channel chan proto.Serializable
	lastAttempt    time.Time
	lastSend       time.Time
	lastReceived   time.Time
	outgoingOrder  []proto.Serializable
	lastServer     string
}

func CreateServerConnection(serverChan chan proto.Serializable) ServerConnection {
	return ServerConnection{buffer: make([]byte, 200), status: Disconnected, packet_channel: serverChan,
		lastAttempt: time.Time{}, lastSend: time.Time{}, outgoingOrder: make([]proto.Serializable, 0)}
}

func (sc *ServerConnection) Stop() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	if sc.status != Disconnected {
		sc.connection.Close()
	}
	sc.status = Disconnected
}

func (sc *ServerConnection) Start(address string) {
	sc.mutex.Lock()
	if sc.status != Disconnected {
		sc.mutex.Unlock()
		return
	}

	sc.status = Connecting
	sc.lastServer = address
	sc.mutex.Unlock()

	connection, err := net.Dial("tcp", address)
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
		sc.mutex.Lock()
		sc.status = Disconnected
		sc.mutex.Unlock()
		return
	}

	pc := proto.PacketCombiner{}
	for {
		ph, bytes, error := pc.Read(connection)

		if error != nil {
			fmt.Printf("Can not read bytes from server %v", error)
			break
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
				sc.packet_channel <- &bc
				fmt.Println("Receive message from server", string(bc))
			}
		case proto.OP_SERVERSTATUS:
			ss := proto.Status{}
			ss.Get(&sb)
			if sb.Error() == nil {
				sc.packet_channel <- &ss
				fmt.Println("Server status files:", ss.FilesCount, "users", ss.UsersCount)
			}
		case proto.OP_IDCHANGE:
			idc := proto.IdChange{}
			idc.Get(&sb)
			if sb.Error() == nil {
				fmt.Println("Server id change", idc.ClientId)
				sc.packet_channel <- &idc
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
				sc.packet_channel <- &p
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
			break
		}

		// finalize server connection status
		sc.mutex.Lock()
		sc.status = Connected
		sc.lastSend = time.Time{}
		sc.lastReceived = time.Now()
		if len(sc.outgoingOrder) > 0 {
			sc.outgoingOrder = sc.outgoingOrder[1:]
		}

		fmt.Printf("Data received. Outgoing order size %d\n", len(sc.outgoingOrder))
		sc.mutex.Unlock()

		go sc.Send()
	}

	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.status = Disconnected
	sc.connection.Close()
}

func (sc *ServerConnection) Search(s string) error {
	parsed, err := proto.BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, s)
	if err != nil {
		return err
	}

	req, err := proto.PackRequest(parsed)
	if err != nil {
		return err
	}

	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	if sc.status == Connected {
		sc.outgoingOrder = append(sc.outgoingOrder, &req)
		go sc.Send()
	}
	return nil
}

func (sc *ServerConnection) Tick(t time.Time) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	if !sc.lastSend.IsZero() && t.Sub(sc.lastSend).Seconds() > 10 {
		fmt.Printf("ServerConnection Tick. Outgoing order size: %d", len(sc.outgoingOrder))
		if len(sc.outgoingOrder) > 0 {
			sc.outgoingOrder = sc.outgoingOrder[1:]
		}
		sc.lastSend = time.Time{}
		go sc.Send()
	}
}

func (sc *ServerConnection) Send() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	// check the connection is ready to send data
	if sc.status != Connected || !sc.lastSend.IsZero() {
		return
	}

	if len(sc.outgoingOrder) > 0 {
		_, err := Send(sc.connection, sc.outgoingOrder[0])

		if err != nil {
			defer sc.Stop()
		} else {
			sc.lastSend = time.Now()
		}
	}
}

func (sc ServerConnection) Status() int {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	return sc.status
}

func (sc ServerConnection) IsConnected() bool {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	return sc.status == Connected
}
