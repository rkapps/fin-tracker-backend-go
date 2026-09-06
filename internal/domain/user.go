package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

type ctxToken int

const (
	USER_COL                = "user"
	UserContextUID ctxToken = iota
)

type User struct {
	ID                   string          `json:"id" bson:"id"`
	CurrencyCode         string          `json:"currency"`
	Country              string          `json:"country"`
	MinAccountBalance    decimal.Decimal `json:"minAccountBalance" bson:"minAccountBalance"`
	ShowInactiveAccounts bool            `json:"showInactiveAccounts" bson:"showInactiveAccounts"`
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
