package books

import "errors"

var ErrorNotEnoughData error = errors.New("We don't have enough data to add book")