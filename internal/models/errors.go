package models

import "fmt"

type Error struct {
	Message string `json:"message"`
}

func Err(msg string) Error {
	return Error{
		Message: msg,
	}
}

func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}
