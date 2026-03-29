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

type AppSession struct {
	ID                string
	UserID            string
	WalletID          string
	PrincipalKind     PrincipalKind
	Token             string
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	CreatedAt         time.Time
	LastUsedAt        time.Time
	RevokedAt         *time.Time
}
