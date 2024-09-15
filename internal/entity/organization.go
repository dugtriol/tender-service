package entity

import (
	"time"
)

type Organization struct {
	Id               string    `db:"id"`
	Name             string    `db:"name"`
	Description      string    `db:"description"`
	OrganizationType string    `db:"organization_type"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
