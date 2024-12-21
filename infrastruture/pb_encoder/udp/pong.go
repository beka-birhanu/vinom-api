package udppb

import "github.com/beka-birhanu/vinom-api/service/i"

var _ i.PongRecord = &Pong{}

func (p *Pong) SetPingSentAt(t int64) {
	p.PingSentAt = t
}

func (p *Pong) SetReceivedAt(t int64) {
	p.ReceivedAt = t
}

func (p *Pong) SetSentAt(t int64) {
	p.SentAt = t
}
