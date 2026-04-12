package domain

import "context"

type ctxToken int

const (
	USER_COL                = "user"
	UserContextUID ctxToken = iota
)

type User struct {
	ID       string `json:"id" bson:"id"`
	Currency string `json:"currency"`
	Country  string `json:"country"`
}

func (u *User) Id() string {
	return u.ID
}

func (u *User) CollectionName() string {
	return USER_COL
}

func UserFromCtx(ctx context.Context) User {
	return ctx.Value(UserContextUID).(User)
}
