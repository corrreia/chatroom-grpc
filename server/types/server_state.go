package types

import "errors"

//server state interface
type ServerStater interface {
	//user management
	AddUser(user User) error
	RemoveUser(user User) error
	IsUserRegistered(user string) bool

	//user list
	GetUserList() []User
	GetConnectedUserList() []User
	GetBannedUserList() []User
	GetAdminUserList() []User

	//user info
	GetUserByUsername(user string) User
	GetUserByToken(token string) User
	GetUserById(id string) User

	//server info
	GetServerPassword() string
	GetMaxClients() int
	GetCaPath() string
	GetCertPath() string
	GetKeyPath() string
	GetCurrentClients() int

	//server state
	SetServerPassword(password string) error
	SetMaxClients(max int) error
	SetCaPath(path string) error
	SetCertPath(path string) error
	SetKeyPath(path string) error
}

//server state struct
type ServerState struct {
	Users map[string]User //map of users id: user

	serverPass string
	maxClients int

	caPath string
	certPath string
	keyPath string
}

//user interface is present in user.go

//server state interface implementation
func (s *ServerState) AddUser(user User) error {
	if s.IsUserRegistered(user.GetUsername()) {
		return errors.New("user already registered")
	}

	s.Users[user.GetId()] = user
	return nil
}

func (s *ServerState) RemoveUser(user User) error {
	if !s.IsUserRegistered(user.GetUsername()) {
		return errors.New("user not registered")
	}

	delete(s.Users, user.GetId())
	return nil
}	

func (s *ServerState) IsUserRegistered(user string) bool {
	_, ok := s.Users[user]
	return ok
}

func (s *ServerState) GetUserList() []User {
	var users []User
	for _, user := range s.Users {
		users = append(users, user)
	}

	return users
}

func (s *ServerState) GetConnectedUserList() []User {
	var users []User
	for _, user := range s.Users {
		if user.IsConnected() {
			users = append(users, user)
		}
	}

	return users
}

func (s *ServerState) GetBannedUserList() []User {
	var users []User
	for _, user := range s.Users {
		if user.IsBanned() {
			users = append(users, user)
		}
	}

	return users
}

func (s *ServerState) GetAdminUserList() []User {
	var users []User
	for _, user := range s.Users {
		if user.IsAdmin() {
			users = append(users, user)
		}
	}

	return users
}

func (s *ServerState) GetUserById(id string) User {
	return s.Users[id]
}

func (s *ServerState) GetUserByUsername(user string) User {
	for _, u := range s.Users {
		if u.GetUsername() == user {
			return u
		}
	}

	return User{}
}

func (s *ServerState) GetUserByToken(token string) User {
	for _, u := range s.Users {
		if u.GetToken() == token {
			return u
		}
	}

	return User{}
}

func (s *ServerState) GetServerPassword() string {
	return s.serverPass
}

func (s *ServerState) GetMaxClients() int {
	return s.maxClients
}

func (s *ServerState) SetServerPassword(password string) error {
	s.serverPass = password
	return nil
}

func (s *ServerState) SetMaxClients(max int) error {
	s.maxClients = max
	return nil
}

func (s *ServerState) GetCaPath() string {
	return s.caPath
}

func (s *ServerState) GetCertPath() string {
	return s.certPath
}

func (s *ServerState) GetKeyPath() string {
	return s.keyPath
}

func (s *ServerState) SetCaPath(path string) error {
	s.caPath = path
	return nil
}

func (s *ServerState) SetCertPath(path string) error {
	s.certPath = path
	return nil
}

func (s *ServerState) SetKeyPath(path string) error {
	s.keyPath = path
	return nil
}

func (s *ServerState) GetCurrentClients() int {
	return len(s.GetConnectedUserList())
}

//server state constructor
func NewServerState() *ServerState {
	s := &ServerState{
		Users: make(map[string]User),
	}

	return s
}




