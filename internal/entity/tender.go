package entity

import "time"

type Tender struct {
	Id              string    `db:"id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	ServiceType     string    `db:"service_type"`
	Status          string    `db:"status"`
	OrganizationId  string    `db:"organization_id"`
	Version         int       `db:"version"`
	CreatedAt       time.Time `db:"created_at"`
	CreatorUsername string    `db:"creator_username"`
}
