package i

import (
	dmn "github.com/beka-birhanu/vinom-api/domain"
	"github.com/google/uuid"
)

// PlayerAuthenticator an interface for authenticating the client token
type PlayerAuthenticator interface {
	Authenticate([]byte) (uuid.UUID, error)
}

type Authenticator interface {
	Register(string, string) error
	SignIn(string, string) (*dmn.User, string, error)
}
