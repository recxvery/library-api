package books

import (
	"sync"

	"github.com/google/uuid"
)

type Lib struct {
	mtx   sync.RWMutex
	books map[string]*Book
}

func NewLib() *Lib {
	return &Lib{
		books: make(map[string]*Book),
	}
}

func (l *Lib) AddBook(book Book) Book {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	
	id := uuid.New().String()
	book.ID = id
	l.books[id] = &book
	
	return *l.books[id]
}

// func (l *Lib) GetBookInfo(id string) (Book, error) {
// 	l.mtx.RLock()
// 	defer l.mtx.RUnlock()
	
// 	book, ok := l.books[id]
// 	if !ok {
// 		return Book{}, ErrorThereIsNoBook
// 	}
	
// 	return *book, nil
// }

func (l *Lib) GetAllBooks() map[string]Book {
	tmp := make(map[string]Book)

	l.mtx.RLock()
	defer l.mtx.RUnlock()

	for id, book := range l.books {
		tmp[id] = *book
	}

	return tmp
}

func(l *Lib) FilterBooks(author string, read *bool) map[string]Book {
	l.mtx.RLock()
	defer l.mtx.RUnlock()

	tmp := make(map[string]Book)

	for id, book := range l.books {

		if author != "" && book.Author != author {
			continue
		}

		if read != nil &&  book.IsRead != *read { //если состояние задано и если не совпадает параметру пропускаем книгу
			continue
		}

		tmp[id] = *book
	}

	return tmp
}
 
func (l *Lib) RemoveBook(id string) (Book, error) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	book, ok := l.books[id]
	if !ok {
		return Book{}, ErrorNotEnoughData
	}

	delete(l.books, id)

	return *book, nil
}

// func (l *Lib) ReadSome(id string) (Book, error) {
// 	l.mtx.Lock()
// 	defer l.mtx.Unlock()

// 	book, ok := l.books[id]
// 	if !ok {
// 		return Book{}, ErrorThereIsNoBook
// 	}

// 	book.Read()

// 	return *book, nil
// }

	// func (l *Lib) GetBooksByAuthor(author string) map[string]Book {
	// 	tmp := make(map[string]Book)
	
	// 	l.mtx.RLock()
	// 	defer l.mtx.RUnlock()
	
	// 	for id, book := range l.books {
	// 		if book.Author == author {
	// 			tmp[id] = *book
	// 		}
	// 	}
	
	// 	return tmp
	// }
	
	// func (l *Lib) GetReadTrueBooks() map[string]Book {
	// 	tmp := make(map[string]Book)
	
	// 	l.mtx.RLock()
	// 	defer l.mtx.RUnlock()
	
	// 	for id, book := range l.books {
	// 		if book.IsRead == true {
	// 			tmp[id] = *book
	// 		}
	// 	}
	
	// 	return tmp
	// }
	
	
	// func (l *Lib) GetUnreadBooks() map[string]Book {
	// 	tmp := make(map[string]Book)
	
	// 	l.mtx.RLock()
	// 	defer l.mtx.RUnlock()
	
	// 	for id, book := range l.books {
	// 		if book.IsRead == false {
	// 			tmp[id] = *book
	// 		}
	// 	}
	
	// 	return tmp
	// }