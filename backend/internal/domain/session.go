package domain

import "time"

type DemoSession struct {
	ID         string
	UserID     string
	WalletID   string
	Token      string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastUsedAt time.Time
}
