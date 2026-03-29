package domain

type PrincipalKind string

const (
	PrincipalKindGuest      PrincipalKind = "guest"
	PrincipalKindRegistered PrincipalKind = "registered"
)

type Principal struct {
	Kind        PrincipalKind `json:"kind"`
	UserID      string        `json:"userID"`
	WalletID    string        `json:"walletID"`
	SessionID   string        `json:"sessionID"`
	ClerkUserID string        `json:"clerkUserID"`
}

type RegisteredIdentity struct {
	ClerkUserID string `json:"clerkUserID"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type RegisteredAccount struct {
	UserID      string `json:"userID"`
	WalletID    string `json:"walletID"`
	ClerkUserID string `json:"clerkUserID"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}
