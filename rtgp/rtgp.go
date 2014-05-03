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
	"time"
)

var udpConnsLock chan map[int]*udpConnCount
var connsLock chan map[string]*Conn

func init() {
	udpConnsLock = make(chan map[int]*udpConnCount, 1)
	udpConnsLock <- make(map[int]*udpConnCount)
	connsLock = make(chan map[string]*Conn, 1)
	connsLock <- make(map[string]*Conn)
}

type udpConnCount struct {
	count   uint
	udpConn *net.UDPConn
}

type PayloadType struct {
	Size     int
	Periodic bool
}

type Conn struct {
	payloadTypes       []PayloadType
	udpConn            *net.UDPConn
	udpRAddr           *net.UDPAddr
	sending            chan bool
	tickrateLock       chan int
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

const defaultTickrate = 15

func NewConn(lAddr string, payloadTypes []PayloadType) (*Conn, error) {
	c := new(Conn)
	c.payloadTypes = payloadTypes

	c.sending = make(chan bool, 1)
	c.sending <- false

	c.tickrateLock = make(chan int, 1)
	c.tickrateLock <- defaultTickrate

	var err error
	c.lSessionID, err = generateSessionID()
	if err != nil {
		return nil, err
	}

	udpLAddr, err := net.ResolveUDPAddr("udp", lAddr)
	if err != nil {
		return nil, err
	}
	udpConns := <-udpConnsLock
	_, found := udpConns[udpLAddr.Port]
	if !found {
		udpConn, err := net.ListenUDP("udp", udpLAddr)
		if err != nil {
			return nil, err
		}
		udpConns[udpLAddr.Port] = &udpConnCount{0, udpConn}
		go recvUDP(udpConn)
	}
	udpConns[udpLAddr.Port].count++
	c.udpConn = udpConns[udpLAddr.Port].udpConn
	udpConnsLock <- udpConns

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

func (c *Conn) Close() error {
	<-c.sending
	c.sending <- false

	conns := <-connsLock
	delete(conns, c.udpRAddr.String())
	connsLock <- conns

	udpConns := <-udpConnsLock
	udpConnCount := udpConns[c.LocalPort()]
	udpConnCount.count--
	var err error
	if udpConnCount.count == 0 {
		err = udpConnCount.udpConn.Close()
		delete(udpConns, c.LocalPort())
	}
	udpConnsLock <- udpConns
	return err
}

func (c *Conn) LocalPort() int {
	return c.udpConn.LocalAddr().(*net.UDPAddr).Port
}

func (c *Conn) LocalSessionId() uint32 {
	return c.lSessionID
}

func (c *Conn) SetRemoteAddrAndSessionId(raddr string, id uint32) error {
	sending := <-c.sending
	c.sending <- sending
	if sending {
		return fmt.Errorf("remote address and session id already set")
	}

	udpRAddr, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		return err
	}
	c.udpRAddr = udpRAddr
	c.rSessionID = id
	conns := <-connsLock
	conns[c.udpRAddr.String()] = c
	connsLock <- conns

	sending = <-c.sending
	c.sending <- true
	go sendUDP(c)
	return nil
}

func (c *Conn) SetTickRate(tickrate int) {
	<-c.tickrateLock
	c.tickrateLock <- tickrate
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
		if !c.payloadTypes[p.Type].Periodic {
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

		p.Data = make([]byte, c.payloadTypes[p.Type].Size)
		n, err := data.Read(p.Data)
		if err != nil || n != c.payloadTypes[p.Type].Size {
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
		size := c.payloadTypes[p.Type].Size
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

		size := c.payloadTypes[toSend[i].Type].Size
		data.Write(toSend[i].Data[:size])
	}
	c.toSendLock <- toSend
}

func sendUDP(c *Conn) {
	tickrate := <-c.tickrateLock
	c.tickrateLock <- tickrate
	ticker := newTicker(tickrate)
	for {
		newTickrate := <-c.tickrateLock
		c.tickrateLock <- newTickrate
		if newTickrate != tickrate {
			tickrate = newTickrate
			ticker.Stop()
			ticker = newTicker(tickrate)
		}
		<-ticker.C

		sending := <-c.sending
		c.sending <- sending
		if !sending {
			break
		}

		var data bytes.Buffer

		lSeq := c.writeHeader(&data)
		c.writeToSend(&data, lSeq)
		c.writePeriodicToSend(&data)

		_, err := c.udpConn.WriteToUDP(data.Bytes(), c.udpRAddr)
		if err != nil {
		}
	}
}
