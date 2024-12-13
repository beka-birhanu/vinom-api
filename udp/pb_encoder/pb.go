package pb

import (
	"errors"

	"github.com/beka-birhanu/vinom-api/udp"
	"google.golang.org/protobuf/proto"
)

var _ udp.Encoder = &Protobuf{}

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
func (p *Protobuf) MarshalHandshake(h udp.HandshakeRecord) ([]byte, error) {
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
func (p *Protobuf) MarshalPong(pr udp.PongRecord) ([]byte, error) {
	msg := &Pong{
		PingSentAt: pr.GetPingSentAt(),
		ReceivedAt: pr.GetReceivedAt(),
		SentAt:     pr.GetSentAt(),
	}
	return proto.Marshal(msg)
}

// NewHandshakeRecord implements udp.Encoder.
func (p *Protobuf) NewHandshakeRecord() udp.HandshakeRecord {
	return &Handshake{}
}

// NewPongRecord implements udp.Encoder.
func (p *Protobuf) NewPongRecord() udp.PongRecord {
	return &Pong{}
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
func (p *Protobuf) UnmarshalHandshake(b []byte) (udp.HandshakeRecord, error) {
	h := &Handshake{}
	err := proto.Unmarshal(b, h)
	return h, err
}

// UnmarshalPing implements udp.Encoder.
func (p *Protobuf) UnmarshalPing(b []byte) (udp.PingRecord, error) {
	pi := &Ping{}
	err := proto.Unmarshal(b, pi)
	return pi, err
}
