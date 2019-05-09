package auth

type User struct {
	Name string
}

func CheckLogin(name string, passwd string) (User, error) {
	return User{name}, nil
}
