package domain

type Subscription struct {
	Email string `json:"email"`
	City  string `json:"city"`
}

type ConfirmationEmailInput struct {
	Email            string
	ConfirmationLink string
}
