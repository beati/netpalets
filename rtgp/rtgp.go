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
	"sync"
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

type MsgType struct {
	Size     int
	Reliable bool
}

type msg struct {
	msgType uint16
	data    []byte
}

type periodicMsg struct {
	msgType  uint16
	dataLock chan []byte
}

type reliableMsg struct {
	msg
	seqs []uint32
}

type Conn struct {
	mutex        sync.Mutex
	msgTypes     []MsgType
	udpConn      *net.UDPConn
	udpRAddr     *net.UDPAddr
	sending      bool
	tickrate     uint
	lSessionID   uint32
	rSessionID   uint32
	lSeq         uint32
	rSeq         uint32
	rSeqBits     uint32
	nextMsgID    uint32
	rMsgIDs      map[uint32]struct{}
	periodicMsgs []periodicMsg
	reliableMsgs map[uint32]reliableMsg
	recved       chan struct{}
	recvedMsgs   []msg
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

func NewConn(lAddr string, msgTypes []MsgType, tickrate uint) (*Conn, error) {
	c := new(Conn)
	c.msgTypes = msgTypes
	c.sending = false
	c.tickrate = tickrate
	c.rMsgIDs = make(map[uint32]struct{})
	c.periodicMsgs = make([]periodicMsg, 0)
	c.reliableMsgs = make(map[uint32]reliableMsg)
	c.recved = make(chan struct{}, 1)
	c.recvedMsgs = make([]msg, 0)

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

	return c, nil
}

func (c *Conn) Close() error {
	c.mutex.Lock()
	c.sending = false

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

	c.mutex.Unlock()
	return err
}

func (c *Conn) LocalPort() int {
	return c.udpConn.LocalAddr().(*net.UDPAddr).Port
}

func (c *Conn) LocalSessionId() uint32 {
	return c.lSessionID
}

func (c *Conn) SetRemoteAddrAndSessionId(raddr string, id uint32) error {
	c.mutex.Lock()
	if c.sending {
		c.mutex.Unlock()
		return fmt.Errorf("remote address and session id already set")
	}

	udpRAddr, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		c.mutex.Unlock()
		return err
	}
	c.udpRAddr = udpRAddr
	c.rSessionID = id
	conns := <-connsLock
	conns[c.udpRAddr.String()] = c
	connsLock <- conns

	c.sending = true
	go sendUDP(c)

	c.mutex.Unlock()
	return nil
}

func (c *Conn) SetTickRate(tickrate uint) {
	c.mutex.Lock()
	c.tickrate = tickrate
	c.mutex.Unlock()
}

func (c *Conn) SendPeriodicMsg(msgType uint16, dataLock chan []byte) {
	c.mutex.Lock()
	c.periodicMsgs = append(c.periodicMsgs, periodicMsg{msgType, dataLock})
	c.mutex.Unlock()
}

func (c *Conn) SendMsg(msgType uint16, data []byte, now bool) {
}

func (c *Conn) SendReliableMsg(msgType uint16, data []byte, now bool) {
	seqs := make([]uint32, 0)
	c.mutex.Lock()
	c.reliableMsgs[c.nextMsgID] = reliableMsg{msg{msgType, data}, seqs}
	c.nextMsgID++
	c.mutex.Unlock()
}

func (c *Conn) RecvMsg() (msgType uint16, data []byte) {
	c.mutex.Lock()
	nMsgs := len(c.reliableMsgs)
	c.mutex.Unlock()
	if nMsgs == 0 {
		<-c.recved
	}
	c.mutex.Lock()
	msgType = c.recvedMsgs[len(c.recvedMsgs)-1].msgType
	data = c.recvedMsgs[len(c.recvedMsgs)-1].data
	c.recvedMsgs = c.recvedMsgs[:len(c.recvedMsgs)-1]
	c.mutex.Unlock()
	return
}

type packetHeader struct {
	SessionID uint32
	LSeq      uint32
	RSeq      uint32
	RSeqBits  uint32
}

const maxPacketSize = 1400

func (c *Conn) updateRSeqs(newRSeq uint32) bool {
	if newRSeq < c.rSeq {
		return false
	}
	d := newRSeq - c.rSeq
	c.rSeqBits <<= d
	c.rSeqBits |= 1
	c.rSeq = newRSeq
	return true
}

func acked(seqs []uint32, ackedSeq uint32, ackedSeqBits uint32) bool {
	for _, seq := range seqs {
		var j uint32
		for j = 0; j < 32; j++ {
			if seq == ackedSeq-j && ackedSeqBits|1<<j != 0 {
				return true
			}
		}
	}
	return false
}

func (c *Conn) updateMsgsToSend(ackedSeq uint32, ackedSeqBits uint32) {
	for id, msg := range c.reliableMsgs {
		if acked(msg.seqs, ackedSeq, ackedSeqBits) {
			delete(c.reliableMsgs, id)
		}
	}
}

func (c *Conn) updateRecvedMsgs(data *bytes.Reader) {
	for {
		var m msg
		err := binary.Read(data, binary.LittleEndian, &m.msgType)
		if err != nil {
			break
		}

		alreadyRecved := false
		if c.msgTypes[m.msgType].Reliable {
			var msgID uint32
			err = binary.Read(data, binary.LittleEndian, &msgID)
			if err != nil {
				break
			}

			if _, found := c.rMsgIDs[msgID]; found {
				alreadyRecved = true
			}
			c.rMsgIDs[msgID] = struct{}{}
		}

		m.data = make([]byte, c.msgTypes[m.msgType].Size)
		n, err := data.Read(m.data)
		if err != nil || n != c.msgTypes[m.msgType].Size {
			break
		}

		if !alreadyRecved {
			c.recvedMsgs = append(c.recvedMsgs, m)
			select {
			case c.recved <- struct{}{}:
			default:
			}
		}
	}
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

		c.mutex.Lock()
		if header.SessionID != c.lSessionID {
			//continue
		}

		if !c.updateRSeqs(header.LSeq) {
			continue
		}

		c.updateMsgsToSend(header.RSeq, header.RSeqBits)

		c.updateRecvedMsgs(data)
		c.mutex.Unlock()
	}
}

func newTicker(tickrate uint) *time.Ticker {
	period := time.Duration(1 / float64(tickrate) * float64(time.Second))
	return time.NewTicker(period)
}

func (c *Conn) writeHeader(data *bytes.Buffer) uint32 {
	c.lSeq++
	header := packetHeader{c.rSessionID, c.lSeq, c.rSeq, c.rSeqBits}
	err := binary.Write(data, binary.LittleEndian, header)
	if err != nil {
		log.Fatal(err)
	}
	return c.lSeq
}

func (c *Conn) writePeriodicMsgs(data *bytes.Buffer) {
	for _, msg := range c.periodicMsgs {
		err := binary.Write(data, binary.LittleEndian, msg.msgType)
		if err != nil {
			log.Fatal(err)
		}
		size := c.msgTypes[msg.msgType].Size
		d := <-msg.dataLock
		data.Write(d[:size])
		msg.dataLock <- d
	}
}

func (c *Conn) writeReliableMsgs(data *bytes.Buffer, lSeq uint32) {
	for id, msg := range c.reliableMsgs {
		msg.seqs = append(msg.seqs, lSeq)

		err := binary.Write(data, binary.LittleEndian, msg.msgType)
		if err != nil {
			log.Fatal(err)
		}

		err = binary.Write(data, binary.LittleEndian, id)
		if err != nil {
			log.Fatal(err)
		}

		size := c.msgTypes[msg.msgType].Size
		data.Write(msg.data[:size])
	}
}

func sendUDP(c *Conn) {
	c.mutex.Lock()
	tickrate := c.tickrate
	c.mutex.Unlock()
	ticker := newTicker(tickrate)
	for {
		<-ticker.C

		c.mutex.Lock()
		if !c.sending {
			break
		}

		if tickrate != c.tickrate {
			tickrate = c.tickrate
			ticker.Stop()
			ticker = newTicker(tickrate)
		}

		var data bytes.Buffer

		lSeq := c.writeHeader(&data)
		c.writeReliableMsgs(&data, lSeq)
		c.writePeriodicMsgs(&data)
		fmt.Printf("%d %d %b\n", c.lSeq, c.rSeq, c.rSeqBits)
		c.mutex.Unlock()

		_, err := c.udpConn.WriteToUDP(data.Bytes(), c.udpRAddr)
		if err != nil {
		}
	}
}
