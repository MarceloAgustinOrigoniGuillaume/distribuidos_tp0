package protocol

import (
	"encoding/binary"
	"net"
	"fmt"
)

// PersonBet used by the client
type PersonBet struct {
	Name    string
	Surname string
	Dni     int32
	Birth   string
	Number  int32
}

func (p PersonBet) String() string {
	return fmt.Sprintf(
		"name: %s | surname: %s | dni: %d | birth: %s | number: %d",
		p.Name, p.Surname, p.Dni, p.Birth, p.Number,
	)
}

func (p *PersonBet) MainInfo() string {
	return fmt.Sprintf(
		"dni: %d | numero: %d",
		p.Dni, p.Number,
	)
}


// ClientProtocol encapsulates how to send bets and other data.
type ClientProtocol struct {
	conn   net.Conn
}

// Initialize protocol
func NewClientProtocolFrom(conn net.Conn) *ClientProtocol {
	protocol := &ClientProtocol{
		conn: conn,
	}
	return protocol
}

func NewClientProtocol(servAddr string) (*ClientProtocol, error) {

	conn, err := net.Dial("tcp", servAddr)

	if err != nil {
		return nil, err
	}

	protocol := &ClientProtocol{
		conn: conn,
	}
	return protocol, nil
}



func (protocol *ClientProtocol) sendBytes(data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := protocol.conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

func (protocol *ClientProtocol) sendLen(len uint16) error {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, len)
	return protocol.sendBytes(lenBuf)
}

func (protocol *ClientProtocol) sendInt(value int32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value)) // Have it be big endian.
	return protocol.sendBytes(buf)
}

// Send first the quantity of bytes in the string. Then the string.
func (protocol *ClientProtocol) sendStr(value string) error {
	strBytes := []byte(value)

	if err := protocol.sendLen(uint16(len(strBytes))); err != nil {
		return err
	}

	return protocol.sendBytes(strBytes)
}



// Send bet, means just sending each field for 
func (protocol *ClientProtocol) SendBet(bet *PersonBet) error {
	var err error

	// function to make it compact instead of many ifs.
	send := func(e error) bool {
		err = e
		return err == nil
	}
	
	if (send(protocol.sendStr(bet.Name)) &&
		send(protocol.sendStr(bet.Surname)) &&	
		send(protocol.sendInt(bet.Dni))	&&
		send(protocol.sendStr(bet.Birth)) &&	
		send(protocol.sendInt(bet.Number))){
		return nil
	}

	return err
}

func (protocol *ClientProtocol) Close() error {
	return protocol.conn.Close()
}