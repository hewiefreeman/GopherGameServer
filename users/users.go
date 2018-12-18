// Package users contains all the necessary tools to make and working with Users. A User is a client
// who has successfully logged into the server. You can think of clients who are not attached to a User
// as, for instance, someone in the login screen, but are still connected to the server. A client doesn't
// have to be a User to be able to call your CustomClientActions, so keep that in mind when making them!
package users

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"sync"
)

// User represents a client who has logged into the service. A User can
// be a guest, join/leave/create rooms, and call any client action, including your
// custom client actions. If you are not using the built-in authentication, be aware
// that you will need to make sure any client who has not been authenticated by the server
// can't simply log themselves in through the client API. A User has a lot of useful information,
// so it's highly recommended you look through all the *User methods to get a good understanding
// about everything you can do with them.
//
// WARNING: When you use a *User object in your code, DO NOT dereference it. Instead, there are
// many methods for *User for retrieving any information about them you could possibly need.
// Dereferencing them could cause data races (which will panic and stop the server) in the User
// fields that get locked for synchronizing access.
type User struct {
	name       string
	databaseID int
	isGuest    bool

	//mux LOCKS ALL FIELDS BELOW
	mux     sync.Mutex
	status  int
	friends map[string]*database.Friend
	conns   map[string]*userConn
}

type userConn struct {
	//MUST LOCK clientMux WHEN GETTING/SETTING *user
	clientMux *sync.Mutex
	user      **User

	socket *websocket.Conn
	room   *rooms.Room
	vars   map[string]interface{}
}

var (
	users         map[string]*User = make(map[string]*User)
	usersMux      sync.Mutex
	serverStarted bool = false
	serverName    string
	kickOnLogin   bool = false
	sqlFeatures   bool = false
	rememberMe    bool = false
	multiConnect  bool = false
)

// These represent the four statuses a User could be.
const (
	StatusAvailable = iota // User is available
	StatusInGame           // User is in a game
	StatusIdle             // User is idle
	StatusOffline          // User is offline
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG A USER IN   /////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Login logs a User in to the service.
func Login(userName string, dbID int, autologPass string, isGuest bool, remMe bool, socket *websocket.Conn,
	connUser **User, clientMux *sync.Mutex) (string, error) {
	//REJECT INCORRECT INPUT
	if len(userName) == 0 {
		return "", errors.New("users.Login() requires a user name")
	} else if userName == serverName {
		return "", errors.New("The name '" + userName + "' is unavailable")
	} else if dbID < -1 {
		return "", errors.New("users.Login() requires a database ID (or -1 for no ID)")
	} else if socket == nil {
		return "", errors.New("users.Login() requires a socket")
	}

	//ALWAYS SET A GUEST'S id TO -1
	databaseID := dbID
	if isGuest {
		databaseID = -1
	}

	//MAKE *User IN users MAP & MAKE connID
	var connID string
	var connErr error
	var userExists bool = false
	//
	usersMux.Lock()
	//
	if userOnline, ok := users[userName]; ok {
		userExists = true
		if kickOnLogin {
			//REMOVE USER FROM THEIR CURRENT ROOM IF ANY
			userOnline.mux.Lock()
			for connKey, conn := range userOnline.conns {
				userRoom := (*conn).room
				if userRoom != nil && userRoom.Name() != "" {
					userOnline.mux.Unlock()
					userRoom.RemoveUser(userOnline.name, connKey)
					userOnline.mux.Lock()
				}
				(*(*conn).clientMux).Lock()
				*((*conn).user) = nil
				(*(*conn).clientMux).Unlock()
				//SEND LOGOUT MESSAGE TO CLIENT
				clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, helpers.NewError("", 0))
				(*conn).socket.WriteJSON(clientResp)
			}
			userOnline.mux.Unlock()

			//DELETE FROM users MAP
			delete(users, userName)

			//MAKE connID
			connID = "1"
			userExists = false
		} else if multiConnect {
			//MAKE UNIQUE connID
			for {
				connID, connErr = helpers.GenerateSecureString(5)
				if connErr != nil {
					usersMux.Unlock()
					return "", errors.New("Unexpected login error")
				}
				userOnline.mux.Lock()
				if _, found := (*userOnline).conns[connID]; !found {
					userOnline.mux.Unlock()
					break
				}
				userOnline.mux.Unlock()
			}
		} else {
			usersMux.Unlock()
			return "", errors.New("User '" + userName + "' is already logged in")
		}
	} else if multiConnect {
		//MAKE UNIQUE connID
		connID, connErr = helpers.GenerateSecureString(5)
		if connErr != nil {
			usersMux.Unlock()
			return "", errors.New("Unexpected login error")
		}
	} else {
		//MAKE connID
		connID = "1"
	}
	//MAKE THE userConn
	vars := make(map[string]interface{})
	conn := userConn{socket: socket, room: nil, vars: vars, user: connUser, clientMux: clientMux}
	//FRIENDS OBJECTS
	var friends []map[string]interface{} = []map[string]interface{}{}
	var friendsMap map[string]*database.Friend
	//ADD THE userConn TO THE User OR MAKE THE User
	if userExists {
		(*users[userName]).mux.Lock()
		(*users[userName]).conns[connID] = &conn
		friendsMap = (*users[userName]).friends
		(*users[userName]).mux.Unlock()
		//MAKE FRINDS LIST FOR SERVER RESPONSE
		for _, val := range friendsMap {
			friendEntry := make(map[string]interface{})
			friendEntry["n"] = val.Name()
			friendEntry["rs"] = val.RequestStatus()
			if val.RequestStatus() == database.FriendStatusAccepted {
				//GET THE User STATUS
				if friend, ok := users[val.Name()]; ok {
					friendEntry["s"] = friend.Status()
				} else {
					friendEntry["s"] = StatusOffline
				}
			}
			friends = append(friends, friendEntry)
		}
	} else {
		//GET User's Friend LIST FROM DATABASE
		if dbID != -1 && sqlFeatures {
			var friendsErr error
			friendsMap, friendsErr = database.GetFriends(dbID) // map[string]Friend
			if friendsErr == nil {
				//MAKE FRINDS LIST FOR SERVER RESPONSE
				for _, val := range friendsMap {
					friendEntry := make(map[string]interface{})
					friendEntry["n"] = val.Name()
					friendEntry["rs"] = val.RequestStatus()
					if val.RequestStatus() == database.FriendStatusAccepted {
						//GET THE User STATUS
						if friend, ok := users[val.Name()]; ok {
							friendEntry["s"] = friend.Status()
						} else {
							friendEntry["s"] = StatusOffline
						}
					}
					friends = append(friends, friendEntry)
				}
			}
		}
		conns := make(map[string]*userConn)
		conns[connID] = &conn
		newUser := User{name: userName, databaseID: databaseID, isGuest: isGuest, status: 0,
			friends: friendsMap, conns: conns}
		users[userName] = &newUser
	}
	(*conn.clientMux).Lock()
	*(conn.user) = users[userName]
	(*conn.clientMux).Unlock()
	//
	usersMux.Unlock()

	//SEND ONLINE MESSAGE TO FRIENDS
	statusMessage := make(map[string]interface{})
	statusMessage[helpers.ServerActionFriendStatusChange] = make(map[string]interface{})
	statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = userName
	statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = 0
	for key, val := range friendsMap {
		if val.RequestStatus() == database.FriendStatusAccepted {
			friend, friendErr := Get(key)
			if friendErr == nil {
				friend.mux.Lock()
				for _, friendConn := range friend.conns {
					friendConn.socket.WriteJSON(statusMessage)
				}
				friend.mux.Unlock()
			}
		}
	}

	//SUCCESS, SEND RESPONSE TO CLIENT
	responseVal := make(map[string]interface{})
	responseVal["n"] = userName
	responseVal["f"] = friends
	if rememberMe && len(autologPass) > 0 && remMe {
		responseVal["ai"] = dbID
		responseVal["ap"] = autologPass
	}
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogin, responseVal, helpers.NewError("", 0))
	socket.WriteJSON(clientResp)

	//
	return connID, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AUTOLOG A USER IN   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// AutoLogIn logs a user in automatically with RememberMe and SqlFeatures enabled in ServerSettings.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want the "Remember Me"
// (AKA auto login) feature, enable it in ServerSettings along with the SqlFeatures and corresponding
// options. You can read more about the "Remember Me" login in the project's usage section.
func AutoLogIn(tag string, pass string, newPass string, dbID int, conn *websocket.Conn, connUser **User, clientMux *sync.Mutex) (string, error) {
	//VERIFY AND GET USER NAME FROM DATABASE
	userName, autoLogErr := database.AutoLoginClient(tag, pass, newPass, dbID)
	if autoLogErr != nil {
		return "", autoLogErr
	}
	//
	connID, userErr := Login(userName, dbID, newPass, false, true, conn, connUser, clientMux)
	if userErr != nil {
		return "", userErr
	}
	//
	return connID, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG A USER OUT   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Logout logs a User out from the service. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when logging a User out with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) Logout(connID string) {
	if multiConnect && len(connID) == 0 {
		return
	} else if !multiConnect {
		connID = "1"
	}

	//REMOVE USER FROM THEIR ROOM
	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return
	}
	currRoom := (*u.conns[connID]).room
	if currRoom != nil && currRoom.Name() != "" {
		u.mux.Unlock()
		currRoom.RemoveUser(u.name, connID)
		u.mux.Lock()
	}

	if len(u.conns) == 1 {
		//SEND STATUS CHANGE TO FRIENDS
		statusMessage := make(map[string]interface{})
		statusMessage[helpers.ServerActionFriendStatusChange] = make(map[string]interface{})
		statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = u.name
		statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = StatusOffline
		for key, val := range u.friends {
			if val.RequestStatus() == database.FriendStatusAccepted {
				friend, friendErr := Get(key)
				if friendErr == nil {
					friend.mux.Lock()
					for _, friendConn := range friend.conns {
						(*friendConn).socket.WriteJSON(statusMessage)
					}
					friend.mux.Unlock()
				}
			}
		}
	}
	//LOG USER OUT
	(*u.conns[connID]).clientMux.Lock()
	if *((*u.conns[connID]).user) != nil {
		*((*u.conns[connID]).user) = nil
	}
	(*u.conns[connID]).clientMux.Unlock()
	socket := (*u.conns[connID]).socket
	delete(u.conns, connID)
	if len(u.conns) == 0 {
		// DELETE THE USER IF THERE ARE NO MORE CONNS
		u.mux.Unlock()
		usersMux.Lock()
		delete(users, u.name)
		usersMux.Unlock()
	} else {
		u.mux.Unlock()
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, helpers.NewError("", 0))
	socket.WriteJSON(clientResp)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET A USER   ////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Get finds a logged in User by their name. Returns an error if the User is not online.
func Get(userName string) (*User, error) {
	//REJECT INCORRECT INPUT
	if len(userName) == 0 {
		return &User{}, errors.New("users.Get() requires a user name")
	}

	var user *User
	var ok bool

	usersMux.Lock()
	if user, ok = users[userName]; !ok {
		usersMux.Unlock()
		return &User{}, errors.New("User '" + userName + "' is not logged in")
	}
	usersMux.Unlock()

	//
	return user, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   MAKE A USER JOIN/LEAVE A ROOM   /////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Join makes a User join a Room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when making a User join a Room with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) Join(r *rooms.Room, connID string) error {
	if multiConnect && len(connID) == 0 {
		return errors.New("Must provide a connID when MultiConnect is enabled")
	} else if !multiConnect {
		connID = "1"
	}
	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return errors.New("Invalid connID")
	}
	currRoom := (*u.conns[connID]).room
	if currRoom != nil && currRoom.Name() == r.Name() {
		u.mux.Unlock()
		return errors.New("User '" + u.name + "' is already in room '" + r.Name() + "'")
	} else if currRoom != nil && currRoom.Name() != "" {
		//LEAVE USER'S CURRENT ROOM
		u.mux.Unlock()
		u.Leave(connID)
		u.mux.Lock()
	}
	roomPointer := &((*u.conns[connID]).room)
	varsPointer := &((*u.conns[connID]).vars)
	socketPointer := ((*u.conns[connID]).socket)
	statusPointer := &u.status
	u.mux.Unlock()

	//ADD USER TO DESIGNATED ROOM
	addErr := r.AddUser(u.name, u.databaseID, u.isGuest, socketPointer, roomPointer, varsPointer, statusPointer, &u.mux, connID)
	if addErr != nil {
		return addErr
	}

	//
	return nil
}

// Leave makes a User leave their current room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when making a User leave a Room with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) Leave(connID string) error {
	if multiConnect && len(connID) == 0 {
		return errors.New("Must provide a connID when MultiConnect is enabled")
	} else if !multiConnect {
		connID = "1"
	}

	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return errors.New("Invalid connID")
	}
	currRoom := (*u.conns[connID]).room
	u.mux.Unlock()
	if currRoom != nil && currRoom.Name() != "" {
		removeErr := currRoom.RemoveUser(u.name, connID)
		if removeErr != nil {
			return removeErr
		}
	} else {
		return errors.New("User '" + u.name + "' is not in a room.")
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SET THE STATUS OF A USER   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SetStatus sets the status of a User. Also sends a notification to all the User's Friends (with the request
// status "accepted") that they changed their status.
func (u *User) SetStatus(status int) {

	u.mux.Lock()
	friends := u.friends
	u.status = status
	u.mux.Unlock()

	//SEND STATUS CHANGE MESSAGE TO User's FRIENDS WHOM ARE "ACCEPTED"
	message := make(map[string]interface{})
	message[helpers.ServerActionFriendStatusChange] = make(map[string]interface{})
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = u.name
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = status
	for key, val := range friends {
		if val.RequestStatus() == database.FriendStatusAccepted {
			friend, friendErr := Get(key)
			if friendErr == nil {
				friend.mux.Lock()
				for _, conn := range friend.conns {
					(*conn).socket.WriteJSON(message)
				}
				friend.mux.Unlock()
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   INVITE TO User's PRIVATE ROOM   /////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Invite allows Users to invite other Users to their private Rooms. The inviting User must be in the Room,
// and the Room must be private and owned by the inviting User. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to the inviting User. This must
// be provided when making a User invite another with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) Invite(invUser *User, connID string) error {
	if multiConnect && len(connID) == 0 {
		return errors.New("Must provide a connID when MultiConnect is enabled")
	} else if !multiConnect {
		connID = "1"
	}

	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return errors.New("Invalid connID")
	}
	currRoom := (*u.conns[connID]).room
	u.mux.Unlock()
	rType := rooms.GetRoomTypes()[currRoom.Type()]
	if currRoom == nil || currRoom.Name() == "" {
		return errors.New("The user '" + u.name + "' is not in a room")
	} else if !currRoom.IsPrivate() {
		return errors.New("The room '" + currRoom.Name() + "' is not private")
	} else if currRoom.Owner() != u.name {
		return errors.New("The user '" + u.name + "' is not the owner of the room '" + currRoom.Name() + "'")
	} else if rType.ServerOnly() {
		return errors.New("Only the server can manipulate that type of room")
	}

	//ADD TO INVITE LIST
	addErr := currRoom.AddInvite(invUser.name)
	if addErr != nil {
		return addErr
	}

	//MAKE INVITE MESSAGE
	invMessage := make(map[string]interface{})
	invMessage[helpers.ServerActionRoomInvite] = make(map[string]interface{})
	invMessage[helpers.ServerActionRoomInvite].(map[string]interface{})["u"] = u.name
	invMessage[helpers.ServerActionRoomInvite].(map[string]interface{})["r"] = currRoom.Name()

	//SEND MESSAGE
	invUser.mux.Lock()
	for _, conn := range invUser.conns {
		(*conn).socket.WriteJSON(invMessage)
	}
	invUser.mux.Unlock()

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REVOKE INVITE TO User's PRIVATE ROOM   //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// RevokeInvite revokes the invite to the specified user to their current Room, provided they are online, the Room is private, and this User
// is the owner of the Room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to the inviting User. This must
// be provided when making a User revoke an invite with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) RevokeInvite(revokeUser string, connID string) error {
	if multiConnect && len(connID) == 0 {
		return errors.New("Must provide a connID when MultiConnect is enabled")
	} else if !multiConnect {
		connID = "1"
	}

	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return errors.New("Invalid connID")
	}
	currRoom := (*u.conns[connID]).room
	u.mux.Unlock()
	rType := rooms.GetRoomTypes()[currRoom.Type()]
	if currRoom == nil || currRoom.Name() == "" {
		return errors.New("The user '" + u.name + "' is not in a room")
	} else if !currRoom.IsPrivate() {
		return errors.New("The room '" + currRoom.Name() + "' is not private")
	} else if currRoom.Owner() != u.name {
		return errors.New("The user '" + u.name + "' is not the owner of the room '" + currRoom.Name() + "'")
	} else if rType.ServerOnly() {
		return errors.New("Only the server can manipulate that type of room")
	}

	//REMOVE FROM INVITE LIST
	removeErr := currRoom.RemoveInvite(revokeUser)
	if removeErr != nil {
		return removeErr
	}

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   User ATTRIBUTE READERS   ////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Name gets the name of the User.
func (u *User) Name() string {
	return u.name
}

// DatabaseID gets the database table index of the User.
func (u *User) DatabaseID() int {
	return u.databaseID
}

// Friends gets the Friend list of the User as a map[string]database.Friend where the key string is the friend's
// User name.
func (u *User) Friends() map[string]database.Friend {
	u.mux.Lock()
	friends := make(map[string]database.Friend)
	for key, val := range u.friends {
		friends[key] = *val
	}
	u.mux.Unlock()
	return friends
}

// RoomIn gets the Room that the User is currently in. A nil Room pointer means the User is not in a Room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when getting a User's Room with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) RoomIn(connID string) *rooms.Room {
	if multiConnect && len(connID) == 0 {
		return nil
	} else if !multiConnect {
		connID = "1"
	}
	u.mux.Lock()
	room := (*u.conns[connID]).room
	u.mux.Unlock()
	//
	return room
}

// Status gets the status of the User.
func (u *User) Status() int {
	u.mux.Lock()
	status := u.status
	u.mux.Unlock()
	return status
}

// Socket gets the WebSocket connection of a User. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when getting a User's socket connection with MultiConnect enabled. Otherwise, an empty string can be used.
func (u *User) Socket(connID string) *websocket.Conn {
	if multiConnect && len(connID) == 0 {
		return nil
	} else if !multiConnect {
		connID = "1"
	}
	u.mux.Lock()
	socket := (*u.conns[connID]).socket
	u.mux.Unlock()
	//
	return socket
}

// IsGuest returns true if the User is a guest.
func (u *User) IsGuest() bool {
	return u.isGuest
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// SetServerStarted is for Gopher Game Server internal mechanics only.
func SetServerStarted(val bool) {
	if !serverStarted {
		serverStarted = val
	}
}

// SettingsSet is for Gopher Game Server internal mechanics only.
func SettingsSet(kickDups bool, name string, deleteOnLeave bool, sqlFeat bool, remMe bool, multiConn bool) {
	if !serverStarted {
		kickOnLogin = kickDups
		serverName = name
		sqlFeatures = sqlFeat
		rememberMe = remMe
		multiConnect = multiConn
		rooms.SettingsSet(name, deleteOnLeave, multiConn)
	}
}
