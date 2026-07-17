package service

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidEncryptedBlob = errors.New("invalid encrypted blob")
var ErrInvalidEncryptedItemKey = errors.New("invalid encrypted item key")
var ErrInvalidSharePermission = errors.New("invalid share permission")
var ErrAlreadyShared = errors.New("item already shared with user")
var ErrCannotShareWithSelf = errors.New("cannot share item with yourself")
