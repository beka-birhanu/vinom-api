package i

import "github.com/google/uuid"

// ServerSocketManager manages server-side socket communication and client interactions.
type ServerSocketManager interface {
	// SetClientRequestHandler sets a handler function for processing client requests.
	// The handler takes the client ID, request type, and request data as parameters.
	SetClientRequestHandler(func(uuid.UUID, byte, []byte))

	// SetClientRegisterHandler sets a handler function for registering new clients.
	// The handler takes the client ID as a parameter.
	SetClientRegisterHandler(func(uuid.UUID))
	Stop()
	Serve()
	SetClientAuthenticator(PlayerAuthenticator)
	BroadcastToClients([]uuid.UUID, byte, []byte)
	// GetPublicKey returns the server's public key for secure communication.
	GetPublicKey() []byte

	// GetAddr returns the server's socket address.
	GetAddr() string
}
