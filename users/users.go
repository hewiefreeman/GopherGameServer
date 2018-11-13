// Package users contains all the necessary tools to make and work with Users.
package users

import (
	"errors"
	"sync"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/rooms"
)

// User represents a client who has logged into the service. A User can
// be a guest, join/leave/create rooms, and call any client action, including your
// custom client actions. If you are not using the built-in authentication, be aware
// that you will need to make sure any client who has not been authenticated by the server
// can't simply log themselves in through the client API.
type User struct {
	name       string
	databaseID int
	isGuest    bool

	socket *websocket.Conn

	//mux LOCKS ALL FIELDS BELOW
	mux sync.Mutex
	room *rooms.Room
	status int
	friends map[string]*database.Friend
	vars map[string]interface{}

	onlineMux sync.Mutex
	online bool
}

var (
	users map[string]*User = make(map[string]*User)
	usersMux sync.Mutex
	serverStarted   bool                   = false
	serverName      string
	kickOnLogin     bool = false
	sqlFeatures     bool = false
	rememberMe      bool = false
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
//
// NOTE: If you are using the SQL authentication features, do not use this! Use the client APIs to log in
// your clients, and you can customize your log in process with the database package. Only use this if
// you are making a proper custom authentication for your project.
func Login(userName string, dbID int, autologPass string, isGuest bool, remMe bool, socket *websocket.Conn) (*User, error) {
	//REJECT INCORRECT INPUT
	if len(userName) == 0 {
		return &User{}, errors.New("users.Login() requires a user name")
	} else if userName == serverName {
		return &User{}, errors.New("The name '" + userName + "' is unavailable")
	} else if dbID < -1 {
		return &User{}, errors.New("users.Login() requires a database ID (or -1 for no ID)")
	} else if socket == nil {
		return &User{}, errors.New("users.Login() requires a socket")
	}

	//ALWAYS SET A GUEST'S id TO -1
	databaseID := dbID
	if isGuest {
		databaseID = -1
	}

	//GET User's Friend LIST
	var friends []map[string]interface{} = []map[string]interface{}{}
	var friendsMap map[string]*database.Friend
	if dbID != -1 && sqlFeatures {
		var friendsErr error
		friendsMap, friendsErr = database.GetFriends(dbID) // map[string]Friend
		if friendsErr == nil {
			for _, val := range friendsMap {
				friendEntry := make(map[string]interface{})
				friendEntry["n"] = val.Name()
				friendEntry["rs"] = val.RequestStatus()
				if val.RequestStatus() == database.FriendStatusAccepted {
					//GET THE User STATUS
					user, userErr := Get(val.Name())
					if userErr != nil {
						friendEntry["s"] = StatusOffline
					} else {
						friendEntry["s"] = user.Status()
					}
				}
				friends = append(friends, friendEntry)
			}
		}
	}
	//MAKE *User IN users MAP
	usersMux.Lock();
	if userOnline, ok := users[userName]; ok {
		if kickOnLogin {
			//REMOVE USER FROM THEIR CURRENT ROOM IF ANY
			userRoom := userOnline.RoomIn()
			if userRoom != nil && userOnline.RoomIn().Name() != "" {
				userOnline.RoomIn().RemoveUser(userOnline.name)
			}

			//SEND LOGOUT MESSAGE TO CLIENT
			clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, nil)
			userOnline.socket.WriteJSON(clientResp)

			//LOG USER OUT
			userOnline.onlineMux.Lock()
			userOnline.online = false
			userOnline.onlineMux.Unlock()
			delete(users, userName)
		} else {
			usersMux.Unlock()
			return &User{}, errors.New("User '" + userName + "' is already logged in")
		}
	}
	//ADD THE User TO THE users MAP
	vars := make(map[string]interface{})
	newUser := User{name: userName, databaseID: databaseID, isGuest: isGuest, friends: friendsMap, status: 0, socket: socket,
					room: nil, vars: vars, online: true}
	users[userName] = &newUser
	user := users[userName]
	usersMux.Unlock()

	//SEND ONLINE MESSAGE TO FRIENDS
	message := make(map[string]interface{})
	message[helpers.ServerActionFriendStatusChange] = make(map[string]interface{})
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = userName
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = 0
	for key, val := range friendsMap {
		if val.RequestStatus() == database.FriendStatusAccepted {
			friend, friendErr := Get(key)
			if friendErr == nil {
				friend.socket.WriteJSON(message)
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
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogin, responseVal, nil)
	socket.WriteJSON(clientResp)

	//
	return user, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AUTOLOG A USER IN   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// AutoLogIn logs a user in automatically with RememberMe and SqlFeatures enabled in ServerSettings.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want the "Remember Me"
// (AKA auto login) feature, enable it in ServerSettings along with the SqlFeatures and corresponding
// options. You can read more about the "Remember Me" login in the project's usage section.
func AutoLogIn(tag string, pass string, newPass string, dbID int, conn *websocket.Conn) (*User, error) {
	//VERIFY AND GET USER NAME FROM DATABASE
	userName, autoLogErr := database.AutoLoginClient(tag, pass, newPass, dbID)
	if autoLogErr != nil {
		return &User{}, autoLogErr
	}
	//
	user, userErr := Login(userName, dbID, newPass, false, true, conn)
	if userErr != nil {
		return &User{}, userErr
	}
	//
	return user, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG A USER OUT   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Logout logs a User out from the service.
func (u *User) Logout() {
	//REMOVE USER FROM THEIR ROOM
	currRoom := u.RoomIn()
	if currRoom != nil && currRoom.Name() != "" {
		currRoom.RemoveUser(u.name)
	}

	//GET FRIENDS
	friends := u.Friends();

	//SEND STATUS CHANGE TO FRIENDS
	statusMessage := make(map[string]interface{})
	statusMessage[helpers.ServerActionFriendStatusChange] = make(map[string]interface{})
	statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = u.name
	statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = StatusOffline
	for key, val := range friends {
		if val.RequestStatus() == database.FriendStatusAccepted {
			friend, friendErr := Get(key)
			if friendErr == nil {
				friend.socket.WriteJSON(statusMessage)
			}
		}
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, nil)
	u.socket.WriteJSON(clientResp)

	//LOG USER OUT
	u.onlineMux.Lock()
	u.online = false
	u.onlineMux.Unlock()
	usersMux.Lock()
	delete(users, u.name)
	usersMux.Unlock()
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

// Join makes a User join a Room.
func (u *User) Join(r *rooms.Room) error {
	currRoom := u.RoomIn()
	if currRoom != nil && currRoom.Name() == r.Name() {
		return errors.New("User '" + u.name + "' is already in room '" + r.Name() + "'")
	} else if currRoom == nil || currRoom.Name() != "" {
		//LEAVE USER'S CURRENT ROOM
		u.Leave()
	}

	//ADD USER TO DESIGNATED ROOM
	addErr := r.AddUser(u.name, u.databaseID, u.isGuest, u.socket, &u.room, &u.vars, &u.status, &u.mux)
	if addErr != nil {
		return addErr
	}

	//
	return nil
}

// Leave makes a User leave their current room.
func (u *User) Leave() error {
	currRoom := u.RoomIn()
	if currRoom != nil && currRoom.Name() != "" {
		removeErr := currRoom.RemoveUser(u.name)
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
func (u *User) SetStatus(status int) error {
	friends := u.Friends()
	u.mux.Lock()
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
				friend.socket.WriteJSON(message)
			}
		}
	}

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   INVITE TO User's PRIVATE ROOM   /////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Invite allows Users to invite other Users to their private Rooms. The inviting User must be in the Room,
// and the Room must be private and owned by the inviting User.
func (u *User) Invite(invUser *User, room *rooms.Room) error {
	currRoom := u.RoomIn()
	if currRoom == nil || currRoom.Name() == "" {
		return errors.New("The user '"+u.name+"' is not in the room '"+room.Name()+"'")
	} else if !room.IsPrivate() {
		return errors.New("The room '" + room.Name() + "' is not private")
	} else if room.Owner() != u.name {
		return errors.New("The user '" + u.name + "' is not the owner of the room '" + room.Name() + "'")
	}

	//ADD TO INVITE LIST
	addErr := room.AddInvite(invUser.name)
	if addErr != nil {
		return addErr
	}

	//MAKE INVITE MESSAGE
	invMessage := make(map[string]interface{})
	invMessage[helpers.ServerActionRoomInvite] = make(map[string]interface{})
	invMessage[helpers.ServerActionRoomInvite].(map[string]interface{})["u"] = u.name
	invMessage[helpers.ServerActionRoomInvite].(map[string]interface{})["r"] = room.Name()

	//SEND MESSAGE
	invUser.socket.WriteJSON(invMessage)

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REVOKE INVITE TO User's PRIVATE ROOM   //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// RevokeInvite revokes the invite to the specified user to the specified Room, provided they are online, the Room is private, and this User
// is the owner of the Room.
func (u *User) RevokeInvite(revokeUser string, room *rooms.Room) error {
	currRoom := u.RoomIn()
	if currRoom == nil || currRoom.Name() == "" {
		return errors.New("The user '"+u.name+"' is not in the room '"+room.Name()+"'")
	} else if !room.IsPrivate() {
		return errors.New("The room '" + room.Name() + "' is not private")
	} else if room.Owner() != u.name {
		return errors.New("The user '" + u.name + "' is not the owner of the room '" + room.Name() + "'")
	}

	//REMOVE FROM INVITE LIST
	removeErr := room.RemoveInvite(revokeUser)
	if removeErr != nil {
		return removeErr
	}

	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   KICK A USER   ///////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// KickUser logs a User out by their name. Also used by KickDupOnLogin in ServerSettings.
func KickUser(userName string) error {
	if len(userName) == 0 {
		return errors.New("users.KickUser() requires a user name")
	}
	//
	user, err := Get(userName)
	if err != nil {
		return err
	}
	//
	user.Logout()
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
	friends := make(map[string]database.Friend);
	for key, val := range u.friends {
		friends[key] = *val
	}
	u.mux.Unlock()
	return friends
}

// RoomIn gets the Room that the User is currently in. A nil Room pointer means the User is not in a Room.
func (u *User) RoomIn() *rooms.Room {
	u.mux.Lock()
	room := u.room;
	u.mux.Unlock()
	return room
}

// Status gets the status of the User.
func (u *User) Status() int {
	u.mux.Lock()
	status := u.status;
	u.mux.Unlock()
	return status
}

// Socket gets the WebSocket connection of a User.
func (u *User) Socket() *websocket.Conn {
	return u.socket
}

// IsGuest returns true if the User is a guest.
func (u *User) IsGuest() bool {
	return u.isGuest
}

// IsOnline returns true if the User is online.
func (u *User) IsOnline() bool {
	u.onlineMux.Lock()
	online := u.online
	u.onlineMux.Unlock()
	return online
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
func SettingsSet(kickDups bool, name string, deleteOnLeave bool, sqlFeat bool, remMe bool) {
	if !serverStarted {
		kickOnLogin = kickDups
		serverName = name
		sqlFeatures = sqlFeat
		rememberMe = remMe
		rooms.SettingsSet(name, deleteOnLeave)
	}
}
