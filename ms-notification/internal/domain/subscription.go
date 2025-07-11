package domain

type Subscription struct {
	Email string `json:"email"`
	City  string `json:"city"`
	Token string `json:"token"`
}
