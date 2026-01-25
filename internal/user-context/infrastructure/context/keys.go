package context

type ContextKey string

const (
	UserIDKey   ContextKey = "user_id"
	FamilyIDKey ContextKey = "family_id"
)
