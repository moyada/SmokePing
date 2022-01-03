package monitor

import (
	"github.com/go-ping/ping"
)

type NewPinger struct {
	ping.Pinger

	OnTimeout func(packet *ping.Packet)
}

func (p *NewPinger) ping()  {

}
