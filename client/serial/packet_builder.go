package serial


import (
	"errors"
	"fmt"
)

const HARD_LIMIT = 8 * 1024; // Limit in bytes for batch size. 

const ALL_OK_CODE = 0; // All ok code, batch sent properly.
const FINISH_FLAG = 0; // Finish flag== count of 0 bets

// PacketBuilder handles the logic/constraints on Packets, like the hard limit on the packet size.
type PacketBuilder struct {
	data []byte
	cached []byte
	serializer *ClientSerializer // It is a Struct member to avoid locating heap at every bet.
}

// Initialize PacketBuilder
func NewPacketBuilder() *PacketBuilder {
	builder := &PacketBuilder{
		data: make([]byte, 0, HARD_LIMIT),
		serializer: NewClientSerializer(),
	}
	return builder
}

func (builder *PacketBuilder) reset() {
	builder.data = builder.data[:0]

	if len(builder.cached) > cap(builder.data){ // Should not really happen, unless 1 bet has size > 8kb
		builder.data = append(builder.data, builder.cached[:cap(builder.data)] ...)
		
		builder.cached = builder.cached[cap(builder.data):]
	} else{
		
		builder.data = append(builder.data, builder.cached...)
		builder.cached = builder.cached[:0]
	}	
}
func (builder *PacketBuilder) SendFinish(conn *ClientConnection) error {


	builder.serializer.WriteInt(FINISH_FLAG)
	defer builder.serializer.Clear()	

	// In the future wait for winner or so.
	return conn.Send(builder.serializer)
}

func (builder *PacketBuilder) SendPacket(count int32, conn *ClientConnection) error {

	if len(builder.data) == 0 {
		return errors.New("Send packet failed, empty data")		
	}


	builder.serializer.WriteInt(count)
	defer builder.serializer.Clear()

	// Send len
	err:= conn.Send(builder.serializer)
	if err != nil{
		return err
	}

	// Send bets
	err= conn.SendBytes(builder.data)
	if err != nil{
		return err
	}

	// Wait confirm

	code:= int32(-1)
	code, err = conn.ReadInt()

	if err != nil{
		return fmt.Errorf("Error at read status code: %w",err)
	}

	if code != ALL_OK_CODE {
		msg, err2 := conn.ReadStr()

		if err2 != nil {
			return fmt.Errorf("Error code %d , msg error read failed: %w",code,err2)
		}
		return errors.New(msg)
	}

	builder.reset()

	return err 	
}


func (builder *PacketBuilder) freeSpace() int {
	return cap(builder.data) - len(builder.data)
}

// Send bet, means just sending each field for 
func (builder *PacketBuilder) WriteBet(bet *PersonBet) bool {
	builder.serializer.WriteBet(bet)

	newData:= builder.serializer.GetData()
	defer builder.serializer.Clear()

	if len(newData) <= builder.freeSpace() {
		builder.data = append(builder.data, newData...)
		return true
	} else{
		builder.cached= append(builder.cached, newData...)
		return false	
	}
}