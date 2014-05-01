package rtgp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"os"
	"time"
)

var sendOn = false
var tickrateLock chan int
var payloadTypes []PayloadType
var udpConns map[int]*net.UDPConn
var connsLock chan map[string]*Conn

const defaultTickrate = 15

func init() {
	tickrateLock = make(chan int, 1)
	tickrateLock <- defaultTickrate
	udpConns = make(map[int]*net.UDPConn)
	connsLock = make(chan map[string]*Conn, 1)
	connsLock <- make(map[string]*Conn)
}

type PayloadType struct {
	Size     int
	Periodic bool
}

func RegisterPayloadTypes(pt []PayloadType) {
	payloadTypes = pt
}

type Conn struct {
	udpConn            *net.UDPConn
	udpRAddr           *net.UDPAddr
	lSessionID         uint32
	rSessionID         uint32
	lSeq               chan uint32
	rSeq               chan uint32
	rSeqBitfield       chan uint32
	nextPayloadID      uint32
	rPayloadIDs        []uint32
	periodicToSendLock chan []chan Payload
	toSendLock         chan []ackedPayload
	recvedLock         chan []Payload
}

func generateSessionID() (uint32, error) {
	max := big.NewInt(math.MaxUint32)
	max.Add(max, big.NewInt(1))
	id, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return uint32(id.Uint64()), nil
}

func NewConn(lAddr string) (*Conn, error) {
	c := new(Conn)

	var err error
	c.lSessionID, err = generateSessionID()
	if err != nil {
		return nil, err
	}

	udpLAddr, err := net.ResolveUDPAddr("udp", lAddr)
	if err != nil {
		return nil, err
	}
	udpConn, found := udpConns[udpLAddr.Port]
	if !found {
		udpConn, err = net.ListenUDP("udp", udpLAddr)
		if err != nil {
			return nil, err
		}
		udpConns[udpLAddr.Port] = udpConn
		go recvUDP(udpConn)
	}
	c.udpConn = udpConn

	c.lSeq = make(chan uint32, 1)
	c.lSeq <- 0
	c.rSeq = make(chan uint32, 1)
	c.rSeq <- 0
	c.rSeqBitfield = make(chan uint32, 1)
	c.rSeqBitfield <- 0

	c.periodicToSendLock = make(chan []chan Payload, 1)
	c.periodicToSendLock <- make([]chan Payload, 0)

	c.toSendLock = make(chan []ackedPayload, 1)
	c.toSendLock <- make([]ackedPayload, 0)

	c.recvedLock = make(chan []Payload, 1)
	c.recvedLock <- make([]Payload, 0)

	return c, nil
}

func SetTickRate(tickrate int) {
	<-tickrateLock
	tickrateLock <- tickrate
}

func (c *Conn) LocalPort() int {
	return c.udpConn.LocalAddr().(*net.UDPAddr).Port
}

func (c *Conn) LocalSessionId() uint32 {
	return c.lSessionID
}

func (c *Conn) SetRemoteAddrAndSessionId(raddr string, id uint32) error {
	udpRAddr, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		return err
	}
	c.udpRAddr = udpRAddr
	c.rSessionID = id
	conns := <-connsLock
	conns[c.udpRAddr.String()] = c
	connsLock <- conns
	if !sendOn {
		sendOn = true
		go sendUDP()
	}
	return nil
}

type Payload struct {
	Type uint16
	Data []byte
}

type ackedPayload struct {
	Payload
	seqs []uint32
	id   uint32
}

func (c *Conn) SendPeriodically(p chan Payload) {
	toSend := <-c.periodicToSendLock
	toSend = append(toSend, p)
	c.periodicToSendLock <- toSend
}

func (c *Conn) Send(p Payload) {
	seqs := make([]uint32, 0)
	toSend := <-c.toSendLock
	toSend = append(toSend, ackedPayload{p, seqs, c.nextPayloadID})
	c.toSendLock <- toSend
	c.nextPayloadID++
}

func (c *Conn) Recv() []Payload {
	recved := <-c.recvedLock
	c.recvedLock <- make([]Payload, 0)
	return recved
}

type packetHeader struct {
	SessionID    uint32
	LSeq         uint32
	RSeq         uint32
	RSeqBitfield uint32
}

const maxPacketSize = 1400

func (c *Conn) updateRSeqs(newRSeq uint32) bool {
	oldRSeq := <-c.rSeq
	oldRSeqBitfield := <-c.rSeqBitfield
	if newRSeq < oldRSeq {
		c.rSeq <- oldRSeq
		c.rSeqBitfield <- oldRSeqBitfield
		return false
	}
	rSeq := newRSeq
	rSeqBitfield := oldRSeqBitfield
	d := rSeq - oldRSeq
	rSeqBitfield <<= d
	rSeqBitfield |= 1 << (d - 1)
	c.rSeqBitfield <- rSeqBitfield
	c.rSeq <- rSeq
	return true
}

func acked(seqs []uint32, AckedSeq uint32, AckedSeqBitfield uint32) bool {
	for _, seq := range seqs {
		if seq == AckedSeq {
			return true
		}

		var j uint32
		for j = 0; j < 32; j++ {
			if seq == AckedSeq-(j+1) && AckedSeqBitfield|1<<j != 0 {
				return true
			}
		}
	}
	return false
}

func (c *Conn) updateToSend(AckedSeq uint32, AckedSeqBitfield uint32) {
	toSend := make([]ackedPayload, 0)
	oldToSend := <-c.toSendLock
	for i := range oldToSend {
		if !acked(oldToSend[i].seqs, AckedSeq, AckedSeqBitfield) {
			toSend = append(toSend, oldToSend[i])
		}
	}
	c.toSendLock <- toSend
}

func (c *Conn) updateRecved(data *bytes.Reader) {
	recv := <-c.recvedLock
	rPayloadIDs := make([]uint32, 0)
	for {
		var p Payload
		err := binary.Read(data, binary.LittleEndian, &p.Type)
		if err != nil {
			break
		}

		alreadyRecved := false
		if !payloadTypes[p.Type].Periodic {
			var payloadID uint32
			err = binary.Read(data, binary.LittleEndian, &payloadID)
			if err != nil {
				break
			}

			for _, rID := range c.rPayloadIDs {
				if rID == payloadID {
					alreadyRecved = true
				}
			}
			rPayloadIDs = append(rPayloadIDs, payloadID)
		}

		p.Data = make([]byte, payloadTypes[p.Type].Size)
		n, err := data.Read(p.Data)
		if err != nil || n != payloadTypes[p.Type].Size {
			break
		}

		if !alreadyRecved {
			recv = append(recv, p)
		}
	}
	c.rPayloadIDs = rPayloadIDs
	c.recvedLock <- recv
}

func recvUDP(udpConn *net.UDPConn) {
	packetData := make([]byte, maxPacketSize)
	var header packetHeader

	for {
		n, raddr, err := udpConn.ReadFromUDP(packetData)
		if err != nil {
			break
		}
		conns := <-connsLock
		c, found := conns[raddr.String()]
		connsLock <- conns
		if !found {
			continue
		}

		data := bytes.NewReader(packetData[:n])
		err = binary.Read(data, binary.LittleEndian, &header)
		if err != nil {
			continue
		}

		if header.SessionID != c.lSessionID {
			//continue
		}

		if !c.updateRSeqs(header.LSeq) {
			continue
		}

		c.updateToSend(header.RSeq, header.RSeqBitfield)

		c.updateRecved(data)
	}

	fmt.Fprintln(os.Stderr, "close connection")
}

func newTicker(tickrate int) *time.Ticker {
	period := time.Duration(1 / float64(tickrate) * float64(time.Second))
	return time.NewTicker(period)
}

func (c *Conn) writeHeader(data *bytes.Buffer) uint32 {
	lSeq := <-c.lSeq
	c.lSeq <- lSeq + 1
	rSeq := <-c.rSeq
	rSeqBitfield := <-c.rSeqBitfield
	c.rSeqBitfield <- rSeqBitfield
	c.rSeq <- rSeq
	header := packetHeader{c.rSessionID, lSeq, rSeq, rSeqBitfield}
	err := binary.Write(data, binary.LittleEndian, header)
	if err != nil {
		log.Fatal(err)
	}
	return lSeq
}

func (c *Conn) writePeriodicToSend(data *bytes.Buffer) {
	toSend := <-c.periodicToSendLock
	for _, pLock := range toSend {
		p := <-pLock
		err := binary.Write(data, binary.LittleEndian, p.Type)
		if err != nil {
			log.Fatal(err)
		}
		size := payloadTypes[p.Type].Size
		data.Write(p.Data[:size])
		pLock <- p
	}
	c.periodicToSendLock <- toSend
}

func (c *Conn) writeToSend(data *bytes.Buffer, lSeq uint32) {
	toSend := <-c.toSendLock
	for i := range toSend {
		toSend[i].seqs = append(toSend[i].seqs, lSeq)

		err := binary.Write(data, binary.LittleEndian, toSend[i].Type)
		if err != nil {
			log.Fatal(err)
		}

		err = binary.Write(data, binary.LittleEndian, toSend[i].id)
		if err != nil {
			log.Fatal(err)
		}

		size := payloadTypes[toSend[i].Type].Size
		data.Write(toSend[i].Data[:size])
	}
	c.toSendLock <- toSend
}

func sendUDP() {
	tickrate := <-tickrateLock
	tickrateLock <- tickrate
	ticker := newTicker(tickrate)
	for {
		newTickrate := <-tickrateLock
		tickrateLock <- newTickrate
		if newTickrate != tickrate {
			tickrate = newTickrate
			ticker.Stop()
			ticker = newTicker(tickrate)
		}
		<-ticker.C

		conns := <-connsLock
		if len(conns) == 0 {
			sendOn = false
			break
		}

		for _, c := range conns {
			var data bytes.Buffer

			lSeq := c.writeHeader(&data)
			c.writeToSend(&data, lSeq)
			c.writePeriodicToSend(&data)

			_, err := c.udpConn.WriteToUDP(data.Bytes(), c.udpRAddr)
			if err != nil {
			}
		}
		connsLock <- conns
	}
}
