package domain

import "time"

type DemoSession struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userID"`
	WalletID   string    `json:"walletID"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}

type AppSession struct {
	ID                string        `json:"id"`
	UserID            string        `json:"userID"`
	WalletID          string        `json:"walletID"`
	PrincipalKind     PrincipalKind `json:"principalKind"`
	Token             string        `json:"token"`
	IdleExpiresAt     time.Time     `json:"idleExpiresAt"`
	AbsoluteExpiresAt time.Time     `json:"absoluteExpiresAt"`
	CreatedAt         time.Time     `json:"createdAt"`
	LastUsedAt        time.Time     `json:"lastUsedAt"`
	RevokedAt         *time.Time    `json:"revokedAt"`
}
