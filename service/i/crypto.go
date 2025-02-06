package i

// The symmetric cryptography interface
// This type of cryptography uses to encrypt messages for the client & decrypt the client packets after a success handshake
type Symmetric interface {
	Encrypt(p []byte, key []byte) ([]byte, error)
	Decrypt(c []byte, key []byte) ([]byte, error)
}

// The asymmetric cryptography interface
// This type of cryptography uses to decrypt the encrypted message from the client in the handshaking process
type Asymmetric interface {
	Encrypt(p []byte, key []byte) ([]byte, error)
	Decrypt(c []byte) ([]byte, error)
	GetPublicKey() []byte
}

type HMAC interface {
	Sign([]byte, ...[]byte) []byte
	Compare([]byte, []byte) bool
}
