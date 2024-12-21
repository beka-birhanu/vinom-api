package i

type HandshakeRecord interface {
	GetSessionId() []byte
	SetSessionId([]byte)
	GetRandom() []byte
	SetRandom([]byte)
	GetCookie() []byte
	SetCookie([]byte)
	GetToken() []byte
	SetToken([]byte)
	GetKey() []byte
	SetKey([]byte)
	GetTimestamp() int64
	SetTimestamp(int64)
}

type PingRecord interface {
	GetSentAt() int64
	SetSentAt(int64)
}

type PongRecord interface {
	GetPingSentAt() int64
	SetPingSentAt(int64)
	GetReceivedAt() int64
	SetReceivedAt(int64)
	GetSentAt() int64
	SetSentAt(int64)
}

type SocketEncoder interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error

	NewHandshakeRecord() HandshakeRecord
	MarshalHandshake(HandshakeRecord) ([]byte, error)
	UnmarshalHandshake([]byte) (HandshakeRecord, error)

	UnmarshalPing([]byte) (PingRecord, error)
	UnmarshalPong([]byte) (PongRecord, error)
	NewPongRecord() PongRecord
	NewPingRecord() PingRecord
	MarshalPong(PongRecord) ([]byte, error)
	MarshalPing(PingRecord) ([]byte, error)
}
