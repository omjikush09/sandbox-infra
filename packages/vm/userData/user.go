package userdata

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/omjikush09/sandboxing-infra/packages/vm/start"
)

type User struct {
	ID   string
	VM   *start.VM
	Port string
}

func createNewUesr() *User {
	id, _ := uuid.NewV7()
	return &User{
		ID: id.String(),
	}
}

func (u *User) UpdatePort(port string) error {
	if u.Port != "" {
		return fmt.Errorf("Port already open")
	}

	u.Port = port
	return nil

}

func (u *User) UpadateUser(vm *start.VM) {

	if vm != nil {
		u.VM = vm
	}

}

type UserData struct {
	mu   sync.Mutex
	data map[string]*User
}

func (ud *UserData) AddUser() *User {
	ud.mu.Lock()
	defer ud.mu.Unlock()
	user := createNewUesr()
	ud.data[user.ID] = user
	return user
}

func (ud *UserData) GetUser(id string) (*User, error) {
	ud.mu.Lock()
	defer ud.mu.Unlock()

	user, ok := ud.data[id]
	if !ok {
		return nil, fmt.Errorf("User not found")
	}
	return user, nil
}
