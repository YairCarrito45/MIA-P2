package stores

type AuthStore struct {
	IsLoggedIn  bool
	Username    string
	Password    string
	PartitionID string
	UserID      int
	GroupID     int
}
var Auth = &AuthStore{
	IsLoggedIn:  false,
	Username:    "",
	Password:    "",
	PartitionID: "",
	UserID:      -1,
	GroupID:     -1,
}

func (a *AuthStore) Login(username, password, partitionID string, uid, gid int) {
	a.IsLoggedIn = true
	a.Username = username
	a.Password = password
	a.PartitionID = partitionID
	a.UserID = uid
	a.GroupID = gid
}

func (a *AuthStore) Logout() {
	a.IsLoggedIn = false
	a.Username = ""
	a.Password = ""
	a.PartitionID = ""
	a.UserID = -1
	a.GroupID = -1
}

func (a *AuthStore) IsAuthenticated() bool {
	return a.IsLoggedIn
}

func (a *AuthStore) GetCurrentUser() (string, string, string) {
	return a.Username, a.Password, a.PartitionID
}

func (a *AuthStore) GetPartitionID() string {
	return a.PartitionID
}