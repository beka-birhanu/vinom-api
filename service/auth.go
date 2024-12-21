package service

import (
	"errors"

	dmn "github.com/beka-birhanu/vinom-api/domain"
	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/google/uuid"
)

type Auth struct {
	userRepo  i.UserRepo
	tokenizer i.Tokenizer
}

func (a *Auth) Register(username, password string) error {
	userConfig := dmn.UserConfig{
		ID:            uuid.New(),
		Username:      username,
		PlainPassword: password,
	}

	user, err := dmn.NewUser(userConfig)
	if err != nil {
		return err
	}

	err = a.userRepo.Save(user)
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) SignIn(username, password string) (string, error) {
	user, err := a.userRepo.ByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	if !user.VerifyPassword(password) {
		return "", errors.New("invalid username or password")
	}

	token, err := a.tokenizer.Generate(map[string]interface{}{
		"userID":   user.ID,
		"username": user.Username,
	}, 24*60*60)

	return token, err
}
