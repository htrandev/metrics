package repository

import "errors"

var (
	// ErrNotFound возвращается когда метрика не найдена в хранилище.
	ErrNotFound = errors.New("not found in storage")
)
