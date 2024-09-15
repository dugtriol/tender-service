package entity

import "time"

type Bid struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Status      string    `db:"status"`
	TenderId    string    `db:"tender_id"`
	AuthorType  string    `db:"author_type"`
	AuthorId    string    `db:"author_id"`
	Version     int       `db:"version"`
	CreatedAt   time.Time `db:"created_at"`
}
