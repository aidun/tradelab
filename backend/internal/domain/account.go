package domain

type PrincipalKind string

const (
	PrincipalKindGuest      PrincipalKind = "guest"
	PrincipalKindRegistered PrincipalKind = "registered"
)

type Principal struct {
	Kind        PrincipalKind
	UserID      string
	WalletID    string
	SessionID   string
	ClerkUserID string
}

type RegisteredIdentity struct {
	ClerkUserID string
	Email       string
	DisplayName string
}

type RegisteredAccount struct {
	UserID      string
	WalletID    string
	ClerkUserID string
	Email       string
	DisplayName string
}
