package userdata

import "sync"

var userData = UserData{
	mu:   sync.Mutex{},
	data: make(map[string]*User),
}

func CreateUser() *User {
	return userData.AddUser()
}

func GetUser(userId string) (*User, error) {
	return userData.GetUser(userId)
}
