package books

import "errors"

var ErrorNotEnoughData error = errors.New("We don't have enough data to add book")
var ErrorThereIsNoBook error = errors.New("This book doesn't exist")