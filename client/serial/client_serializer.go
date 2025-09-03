package serial

import (
	"encoding/binary"
)


// ClientSerializer encapsulates how to Write bets and other data.
type ClientSerializer struct {
	data   []byte
}

// Initialize serializer
func NewClientSerializer() *ClientSerializer {
	serializer := &ClientSerializer{
		data: make([]byte, 0, 128),
	}
	return serializer
}

func (serializer *ClientSerializer) GetData() []byte {
	return serializer.data
}


func (serializer *ClientSerializer) Clear() {
	serializer.data = serializer.data[:0]
}


func (serializer *ClientSerializer) WriteBytes(data []byte) {
	serializer.data= append(serializer.data, data ...)
}

func (serializer *ClientSerializer) WriteLen(len uint16) {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, len)
	serializer.WriteBytes(lenBuf)
}

func (serializer *ClientSerializer) WriteInt(value int32) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value)) // Have it be big endian.
	serializer.WriteBytes(buf)
}

// Write first the quantity of bytes in the string. Then the string.
func (serializer *ClientSerializer) WriteStr(value string) {
	strBytes := []byte(value)
	serializer.WriteLen(uint16(len(strBytes)))
	serializer.WriteBytes(strBytes)
}



// Write bet, means just Writing each field 
func (serializer *ClientSerializer) WriteBet(bet *PersonBet) {
	serializer.WriteStr(bet.Name)
	serializer.WriteStr(bet.Surname)	
	serializer.WriteInt(bet.Dni)
	serializer.WriteStr(bet.Birth)	
	serializer.WriteInt(bet.Number)
}