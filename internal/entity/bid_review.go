package entity

import "time"

type BidReview struct {
	Id          int       `db:"id"`
	Description string    `db:"description"`
	BidId       string    `db:"bid_id"`
	CreatedAt   time.Time `db:"created_at"`
}
