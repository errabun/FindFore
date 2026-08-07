package entity

type Player struct {
	ID             int64
	Name           string
	Phone          string
	Email          string
	Username       string
	PasswordDigest string
	TokenVersion   int32
}

type PlayerWithDetails struct {
	ID       int64
	Name     string
	Phone    string
	Email    string
	Username string
	Friends  []int64
	Events   []int64
}
