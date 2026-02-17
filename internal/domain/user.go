package domain

type ctxToken int

const (
	USER_COL                = "user"
	UserContextUID ctxToken = iota
)

type User struct {
	ID string
}

func (u *User) Id() string {
	return u.ID
}

func (u *User) CollectionName() string {
	return USER_COL
}
