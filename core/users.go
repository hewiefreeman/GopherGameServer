package core

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/database"
	"github.com/hewiefreeman/GopherGameServer/helpers"
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

	//mux lock all items below
	mux     sync.Mutex
	status  int
	friends map[string]*database.Friend
	conns   map[string]*userConn
}

type userConn struct {
	// Must lock *clientMux when using *user
	clientMux *sync.Mutex
	user      **User

	socket *websocket.Conn

	//Must lock user's mux to use below items
	room *Room
	vars map[string]interface{}
}

var (
	users    map[string]*User = make(map[string]*User)
	usersMux sync.Mutex

	// LoginCallback is only for internal Gopher Game Server mechanics.
	LoginCallback func(string, int, map[string]interface{}, map[string]interface{}) bool
	// LogoutCallback is only for internal Gopher Game Server mechanics.
	LogoutCallback func(string, int)
)

// These represent the four statuses a User could be.
const (
	StatusAvailable = iota // User is available
	StatusInGame           // User is in a game
	StatusIdle             // User is idle
	StatusOffline          // User is offline
)

// Error messages
const (
	errorDenied         = "Action was denied"
	errorRequiredName   = "A user name is required"
	errorRequiredID     = "An ID is required"
	errorRequiredSocket = "A socket is required"
	errorNameUnavail    = "Username is unavailable"
	errorUnexpected     = "Unexpected error"
	errorAlreadyLogged  = "User is already logged in"
	errorServerPaused   = "Server is paused"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG A USER IN   /////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Login logs a User in to the service.
func Login(userName string, dbID int, autologPass string, isGuest bool, remMe bool, socket *websocket.Conn,
	connUser **User, clientMux *sync.Mutex) (string, helpers.GopherError) {
	// Verify input
	if serverPaused {
		return "", helpers.NewError(errorServerPaused, helpers.ErrorServerPaused)
	} else if len(userName) == 0 {
		return "", helpers.NewError(errorRequiredName, helpers.ErrorAuthRequiredName)
	} else if userName == serverName {
		return "", helpers.NewError(errorNameUnavail, helpers.ErrorAuthNameUnavail)
	} else if dbID < -1 {
		return "", helpers.NewError(errorRequiredID, helpers.ErrorAuthRequiredID)
	} else if socket == nil {
		return "", helpers.NewError(errorRequiredSocket, helpers.ErrorAuthRequiredSocket)
	}

	// Guests always have -1 databaseID
	databaseID := dbID
	if isGuest {
		databaseID = -1
	}

	// Callback
	if LoginCallback != nil && !LoginCallback(userName, dbID, nil, nil) {
		return "", helpers.NewError(errorDenied, helpers.ErrorActionDenied)
	}

	// Make *User in users & make connID
	var connID string
	var connErr error
	var userExists bool = false
	//
	usersMux.Lock()
	//
	if userOnline, ok := users[userName]; ok {
		userExists = true
		if kickOnLogin {
			// Kick user & remove from room
			userOnline.mux.Lock()
			for connKey, conn := range userOnline.conns {
				userRoom := (*conn).room
				if userRoom != nil && userRoom.Name() != "" {
					userOnline.mux.Unlock()
					userRoom.RemoveUser(userOnline, connKey)
					userOnline.mux.Lock()
				}
				(*(*conn).clientMux).Lock()
				*((*conn).user) = nil
				(*(*conn).clientMux).Unlock()
				// Send logout message to client
				clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, helpers.NoError())
				(*conn).socket.WriteJSON(clientResp)
			}
			userOnline.mux.Unlock()

			// Remove user from users map
			delete(users, userName)

			// Make connID
			connID = "1"
			userExists = false
		} else if multiConnect {
			// Make a unique connID
			for {
				connID, connErr = helpers.GenerateSecureString(5)
				if connErr != nil {
					usersMux.Unlock()
					return "", helpers.NewError(errorUnexpected, helpers.ErrorAuthUnexpected)
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
			return "", helpers.NewError(errorAlreadyLogged, helpers.ErrorAuthAlreadyLogged)
		}
	} else if multiConnect {
		// Make connID
		connID, connErr = helpers.GenerateSecureString(5)
		if connErr != nil {
			usersMux.Unlock()
			return "", helpers.NewError(errorUnexpected, helpers.ErrorAuthUnexpected)
		}
	} else {
		// Make connID
		connID = "1"
	}
	// Make the userConn
	vars := make(map[string]interface{})
	conn := userConn{socket: socket, room: nil, vars: vars, user: connUser, clientMux: clientMux}
	// Make friends objects
	var u *User
	var friends []map[string]interface{}
	var friendsMap map[string]*database.Friend
	// Add the userConn to the User or make new User
	if userExists {
		(*users[userName]).mux.Lock()
		(*users[userName]).conns[connID] = &conn
		friendsMap = (*users[userName]).friends
		(*users[userName]).mux.Unlock()
		// Make friends list for response
		friends = makeFriendsResponse(friendsMap)
	} else {
		// Get friend list from database
		if dbID != -1 && sqlFeatures {
			var friendsErr error
			if friendsMap, friendsErr = database.GetFriends(dbID); friendsErr == nil {
				// Make friends list for response
				friends = makeFriendsResponse(friendsMap)
			}
		}
		conns := map[string]*userConn{
			connID: &conn,
		}
		newUser := User{name: userName, databaseID: databaseID, isGuest: isGuest, status: 0,
			friends: friendsMap, conns: conns}
		u = &newUser
		users[userName] = u
	}
	(*conn.clientMux).Lock()
	*(conn.user) = users[userName]
	(*conn.clientMux).Unlock()
	//
	usersMux.Unlock()

	// Send online message to friends
	statusMessage := map[string]map[string]interface{}{
		helpers.ServerActionFriendStatusChange: {
			"n": userName,
			"s": 0,
		},
	}
	u.sendToFriends(statusMessage)

	// Login success, send response to client
	var responseVal map[string]interface{}
	if rememberMe && len(autologPass) > 0 && remMe {
		responseVal = map[string]interface{}{
			"n": userName,
			"f": friends,
			"ai": dbID,
			"ap": autologPass,
		}
	} else {
		responseVal = map[string]interface{}{
			"n": userName,
			"f": friends,
		}
	}
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogin, responseVal, helpers.NoError())
	socket.WriteJSON(clientResp)

	//
	return connID, helpers.NoError()
}

func makeFriendsResponse(friendsMap map[string]*database.Friend) []map[string]interface{} {
	friends := make([]map[string]interface{}, len(friendsMap), len(friendsMap))
	i := 0;
	for _, val := range friendsMap {
		frs := val.RequestStatus()
		friendEntry := map[string]interface{}{
			"n":  val.Name(),
			"rs": frs,
		}
		if frs == database.FriendStatusAccepted {
			// Get the friend's status
			if friend, ok := users[val.Name()]; ok {
				friendEntry["s"] = friend.Status()
			} else {
				friendEntry["s"] = StatusOffline
			}
		}
		friends[i] = friendEntry
		i++;
	}
	return friends
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AUTOLOG A USER IN   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// AutoLogIn logs a user in automatically with RememberMe and SqlFeatures enabled in ServerSettings.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want the "Remember Me"
// (AKA auto login) feature, enable it in ServerSettings along with the SqlFeatures and corresponding
// options. You can read more about the "Remember Me" login in the project's usage section.
func AutoLogIn(tag string, pass string, newPass string, dbID int, conn *websocket.Conn, connUser **User, clientMux *sync.Mutex) (string, helpers.GopherError) {
	if serverPaused {
		return "", helpers.NewError(errorServerPaused, helpers.ErrorServerPaused)
	}

	// Verify and get user name from database
	userName, autoLogErr := database.AutoLoginClient(tag, pass, newPass, dbID)
	if autoLogErr.ID != 0 {
		return "", autoLogErr
	}
	// Log user in
	connID, userErr := Login(userName, dbID, newPass, false, true, conn, connUser, clientMux)
	if userErr.ID != 0 {
		return "", userErr
	}

	return connID, helpers.NoError()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG/KICK A USER OUT   ///////////////////////////////////////////////////////////////////////////
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

	// Remove user from their room
	u.mux.Lock()
	if _, ok := u.conns[connID]; !ok {
		u.mux.Unlock()
		return
	}
	currRoom := (*u.conns[connID]).room
	if currRoom != nil && currRoom.Name() != "" {
		u.mux.Unlock()
		currRoom.RemoveUser(u, connID)
		u.mux.Lock()
	}

	if len(u.conns) == 1 {
		// Send status change to friends
		statusMessage := map[string]map[string]interface{}{
			helpers.ServerActionFriendStatusChange: {
				"n": u.name,
				"s": StatusOffline,
			},
		}
		u.sendToFriends(statusMessage)
	}
	// Log user out
	(*u.conns[connID]).clientMux.Lock()
	if *((*u.conns[connID]).user) != nil {
		*((*u.conns[connID]).user) = nil
	}
	(*u.conns[connID]).clientMux.Unlock()
	socket := (*u.conns[connID]).socket
	delete(u.conns, connID)
	if len(u.conns) == 0 {
		// Delete user if there are no more conns
		u.mux.Unlock()
		usersMux.Lock()
		delete(users, u.name)
		usersMux.Unlock()
	} else {
		u.mux.Unlock()
	}

	// Send response
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, helpers.NoError())
	socket.WriteJSON(clientResp)

	// Run callback
	if LogoutCallback != nil {
		LogoutCallback(u.Name(), u.DatabaseID())
	}
}

// Kick will log off all connections on this User.
func (u *User) Kick() {
	u.mux.Lock()

	// Send status change message to friends
	statusMessage := map[string]map[string]interface{}{
		helpers.ServerActionFriendStatusChange: {
			"n": u.name,
			"s": StatusOffline,
		},
	}
	u.sendToFriends(statusMessage)

	// Make response
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, helpers.NoError())

	// Go through all connections
	for connID, conn := range u.conns {
		// Remove from room
		currRoom := (*conn).room
		if currRoom != nil && currRoom.Name() != "" {
			u.mux.Unlock()
			currRoom.RemoveUser(u, connID)
			u.mux.Lock()
		}

		// Log connection out
		(*conn).clientMux.Lock()
		if *((*conn).user) != nil {
			*((*conn).user) = nil
		}
		(*conn).clientMux.Unlock()

		// Send response
		(*conn).socket.WriteJSON(clientResp)
	}

	u.mux.Unlock()

	// Remove from users
	usersMux.Lock()
	delete(users, u.name)
	usersMux.Unlock()

	// Run callback
	if LogoutCallback != nil {
		LogoutCallback(u.Name(), u.DatabaseID())
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET A USER   ////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// GetUser finds a logged in User by their name. Returns an error if the User is not online.
func GetUser(userName string) (*User, error) {
	// Verify input
	if len(userName) == 0 {
		return &User{}, errors.New("users.Get() requires a user name")
	} else if serverPaused {
		return &User{}, errors.New(errorServerPaused)
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
func (u *User) Join(r *Room, connID string) error {
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
		// Leave current room
		u.mux.Unlock()
		u.Leave(connID)
		u.mux.Lock()
	}
	u.mux.Unlock()

	// Add user to room
	addErr := r.AddUser(u, connID)
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
		removeErr := currRoom.RemoveUser(u, connID)
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
	u.status = status
	u.mux.Unlock()

	// Send status to friends
	message := map[string]map[string]interface{}{
		helpers.ServerActionFriendStatusChange: {
			"n": u.name,
			"s": status,
		},
	}
	u.sendToFriends(message)
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
	rType := GetRoomTypes()[currRoom.Type()]
	if currRoom == nil || currRoom.Name() == "" {
		return errors.New("The user '" + u.name + "' is not in a room")
	} else if !currRoom.IsPrivate() {
		return errors.New("The room '" + currRoom.Name() + "' is not private")
	} else if currRoom.Owner() != u.name {
		return errors.New("The user '" + u.name + "' is not the owner of the room '" + currRoom.Name() + "'")
	} else if rType.ServerOnly() {
		return errors.New("Only the server can manipulate that type of room")
	}

	// Add to invite list
	addErr := currRoom.AddInvite(invUser.name)
	if addErr != nil {
		return addErr
	}

	// Make response message
	invMessage := map[string]map[string]interface{}{
		helpers.ServerActionRoomInvite: {
			"u": u.name,
			"r": currRoom.Name(),
		},
	}

	// Send response to all connections
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
	rType := GetRoomTypes()[currRoom.Type()]
	if currRoom == nil || currRoom.Name() == "" {
		return errors.New("The user '" + u.name + "' is not in a room")
	} else if !currRoom.IsPrivate() {
		return errors.New("The room '" + currRoom.Name() + "' is not private")
	} else if currRoom.Owner() != u.name {
		return errors.New("The user '" + u.name + "' is not the owner of the room '" + currRoom.Name() + "'")
	} else if rType.ServerOnly() {
		return errors.New("Only the server can manipulate that type of room")
	}

	// Remove from invite list
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

// UserCount returns the number of Users logged into the server.
func UserCount() int {
	usersMux.Lock()
	length := len(users)
	usersMux.Unlock()
	return length
}

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
func (u *User) RoomIn(connID string) *Room {
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

// ConnectionIDs returns a []string of all the User's connection IDs
func (u *User) ConnectionIDs() []string {
	u.mux.Lock()
	ids := make([]string, 0, len(u.conns))
	for id := range u.conns {
		ids = append(ids, id)
	}
	u.mux.Unlock()
	return ids
}