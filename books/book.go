package books

import "time"

type Book struct {
	ID 	string
	Title      string
	Author     string
	Year       int
	PagesCount int
	IsRead     bool
	AddedAt 	time.Time
	ReadAt   *time.Time
}

func (b *Book) Read() {
	b.IsRead = true

	readAtTime := time.Now()
	// *b.ReadAt = readAtTime буквально: *nil = readTime
	b.ReadAt = &readAtTime
}
