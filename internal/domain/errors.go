package domain

import "errors"

var ErrNotFound = errors.New("resource not found")
var ErrInsufficientStock = errors.New("insufficient stock")
