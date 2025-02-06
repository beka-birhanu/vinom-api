package udppb

import (
	"errors"

	"github.com/beka-birhanu/vinom-api/service/i"
	"google.golang.org/protobuf/proto"
)

var _ i.SocketEncoder = &Protobuf{}

var (
	errInvalidProtobufMessage = errors.New("invalid protobuf message")
)

type Protobuf struct{}

// Marshal implements udp.Encoder.
func (p *Protobuf) Marshal(msg interface{}) ([]byte, error) {
	m, ok := msg.(proto.Message)
	if !ok {
		return nil, errInvalidProtobufMessage
	}
	return proto.Marshal(m)
}

// MarshalHandshake implements udp.Encoder.
func (p *Protobuf) MarshalHandshake(h i.HandshakeRecord) ([]byte, error) {
	msg := &Handshake{
		SessionId: h.GetSessionId(),
		Random:    h.GetRandom(),
		Cookie:    h.GetCookie(),
		Token:     h.GetToken(),
		Key:       h.GetKey(),
		Timestamp: h.GetTimestamp(),
	}
	return proto.Marshal(msg)
}

// MarshalPong implements udp.Encoder.
func (p *Protobuf) MarshalPong(pr i.PongRecord) ([]byte, error) {
	msg := &Pong{
		PingSentAt: pr.GetPingSentAt(),
		ReceivedAt: pr.GetReceivedAt(),
		SentAt:     pr.GetSentAt(),
	}
	return proto.Marshal(msg)
}

// MarshalPing implements udp.Encoder.
func (p *Protobuf) MarshalPing(pr i.PingRecord) ([]byte, error) {
	msg := &Ping{
		SentAt: pr.GetSentAt(),
	}
	return proto.Marshal(msg)
}

// NewHandshakeRecord implements udp.Encoder.
func (p *Protobuf) NewHandshakeRecord() i.HandshakeRecord {
	return &Handshake{}
}

// NewPongRecord implements udp.Encoder.
func (p *Protobuf) NewPongRecord() i.PongRecord {
	return &Pong{}
}

// NewPingRecord implements udp.Encoder.
func (p *Protobuf) NewPingRecord() i.PingRecord {
	return &Ping{}
}

// Unmarshal implements udp.Encoder.
func (p *Protobuf) Unmarshal(raw []byte, msg interface{}) error {
	m, ok := msg.(proto.Message)
	if !ok {
		return errInvalidProtobufMessage
	}
	return proto.Unmarshal(raw, m)
}

// UnmarshalHandshake implements udp.Encoder.
func (p *Protobuf) UnmarshalHandshake(b []byte) (i.HandshakeRecord, error) {
	h := &Handshake{}
	err := proto.Unmarshal(b, h)
	return h, err
}

// UnmarshalPing implements udp.Encoder.
func (p *Protobuf) UnmarshalPing(b []byte) (i.PingRecord, error) {
	pi := &Ping{}
	err := proto.Unmarshal(b, pi)
	return pi, err
}

// UnmarshalPing implements udp.Encoder.
func (p *Protobuf) UnmarshalPong(b []byte) (i.PongRecord, error) {
	po := &Pong{}
	err := proto.Unmarshal(b, po)
	return po, err
}
