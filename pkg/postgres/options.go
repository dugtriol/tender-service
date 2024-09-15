package postgres

import "time"

type Option func(db *Database)

func MaxPoolSize(size int) Option {
	return func(c *Database) {
		c.maxPoolSize = size
	}
}

func ConnAttempts(attempts int) Option {
	return func(c *Database) {
		c.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(c *Database) {
		c.connTimeout = timeout
	}
}
