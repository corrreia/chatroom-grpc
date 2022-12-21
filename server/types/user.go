package types

import (
	"github.com/corrreia/chatroom-grpc/utils"
)

//user interface
type Userer interface {
	GetId() string
	GetUsername() string
	GetToken() string
	GetPassword() string
	IsAdmin() bool
	IsBanned() bool
	IsConnected() bool

	SetUsername(name string) error
	SetToken(token string) error
	SetPassword(password string) error
	SetAdmin(admin bool) error
	SetBanned(banned bool) error
	SetConnected(connected bool) error

	RegenerateToken() error
	CheckPassword() bool
}

type User struct {
	id string
	username string
	password string
	token    string

	admin bool
	banned bool	
	connected bool
}

func NewUser(id string, username string, password string) *User {
	return &User{
		id: id,
		username: username,
		password: password,
		token: utils.GenerateToken(),
		admin: false,
		banned: false,
		connected: false,
	}
}

func (u User) GetId() string {
	return u.id
}

func (u *User) GetUsername() string {
	return u.username
}

func (u *User) GetToken() string {
	return u.token
}

func (u *User) GetPassword() string {
	return u.password
}

func (u *User) IsAdmin() bool {
	return u.admin
}

func (u *User) IsBanned() bool {
	return u.banned
}

func (u *User) IsConnected() bool {
	return u.connected
}

func (u *User) SetUsername(name string) error {
	u.username = name
	return nil
}

func (u *User) SetToken(token string) error {
	u.token = token
	return nil
}

func (u *User) SetPassword(password string) error {
	hash, err := utils.GenerateHash(password)
	if err != nil {
		return err
	}
	u.password = hash
	return nil
}

func (u *User) SetAdmin(admin bool) error {
	u.admin = admin
	return nil
}

func (u *User) SetBanned(banned bool) error {
	u.banned = banned
	return nil
}

func (u *User) SetConnected(connected bool) error {
	u.connected = connected
	return nil
}

func (u *User) RegenerateToken() error {
	u.token = utils.GenerateToken()
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return utils.CheckPassword(password, u.password)
}