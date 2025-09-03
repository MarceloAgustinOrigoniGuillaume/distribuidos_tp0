package serial

import (
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