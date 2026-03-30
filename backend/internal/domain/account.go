package domain

type PrincipalKind string

const (
	PrincipalKindGuest      PrincipalKind = "guest"
	PrincipalKindRegistered PrincipalKind = "registered"
)

// Principal represents the authenticated TradeLab actor used for authorization decisions.
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

// RegisteredAccount is the durable application account mapped from a Clerk identity.
type RegisteredAccount struct {
	UserID      string `json:"userID"`
	WalletID    string `json:"walletID"`
	ClerkUserID string `json:"clerkUserID"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}
