package i

import (
	dmn "github.com/beka-birhanu/vinom-api/domain"
)

type Authenticator interface {
	Register(string, string) error
	SignIn(string, string) (*dmn.User, string, error)
}
