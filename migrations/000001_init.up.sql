CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		book_title VARCHAR(1000) NOT NULL,
		book_author VARCHAR(200) NOT NULL,
		publish_year INTEGER NOT NULL,
		pages_count 	INTEGER NOT NULL,
		book_read		BOOLEAN NOT NULL,
		added_at		TIMESTAMP NOT NULL,
		read_at			TIMESTAMP
	);