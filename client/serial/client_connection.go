package serial

import (
	"encoding/binary"
	"net"
)

// ClientConnection handles the short send and short read issues.
// Also handles the read of primitive data for now 
type ClientConnection struct {
	conn   net.Conn
}

// Initialize protocol
func NewClientConnectionFrom(conn net.Conn) *ClientConnection {
	res := &ClientConnection{
		conn: conn,
	}
	return res
}

func NewClientConnection(servAddr string) (*ClientConnection, error) {

	conn, err := net.Dial("tcp", servAddr)

	if err != nil {
		return nil, err
	}

	res := &ClientConnection{
		conn: conn,
	}
	return res, nil
}


func (c *ClientConnection) SendBytes(data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := c.conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}




func (c *ClientConnection) Send(data *ClientSerializer) error {
	return c.SendBytes(data.GetData())
}

func (c *ClientConnection) ReadBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	read := 0
	for read < n {
		nRead, err := c.conn.Read(buf[read:])
		if err != nil {
			return nil, err
		}
		read += nRead
	}
	return buf, nil
}

func (c *ClientConnection) ReadInt() (int32, error) {
	buf, err := c.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf)), nil
}

// ReadU16 reads a 2-byte uint16 (big-endian)
func (c *ClientConnection) ReadU16() (uint16, error) {
	buf, err := c.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf), nil
}

// ReadStr reads a length-prefixed string (2-byte uint16 length)
func (c *ClientConnection) ReadStr() (string, error) {
	length, err := c.ReadU16()
	if err != nil {
		return "", err
	}
	buf, err := c.ReadBytes(int(length))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}


func (c *ClientConnection) Close() error {
	return c.conn.Close()
}