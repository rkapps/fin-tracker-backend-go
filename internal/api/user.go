package api

type ctxToken int

const (
	UserContextUID ctxToken = iota
)

type User struct {
	ID string
}

func (u *User) Id() string {
	return u.ID
}

// SetId sets the unique id for the ticket
func (u *User) SetId() {
}

func (u *User) CollectionName() string {
	return "user"
}
