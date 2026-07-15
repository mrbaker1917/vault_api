package service

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidEncryptedBlob = errors.New("invalid encrypted blob")
