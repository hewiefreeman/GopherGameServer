// Package rooms contains all the necessary tools to make and work with Rooms. A Room represents
// a place on the server where a User can join other Users.
//
// A Room can either be public or private. Private Rooms must be assigned an "owner", which is the name of a User, or the ServerName
// from ServerSettings. The server's name that will be used for ownership of private Rooms can be set with the ServerSettings
// option ServerName when starting the server. Though keep in mind, setting the ServerName in ServerSettings will prevent a User who wants to go by that name
// from logging in. Public Rooms will accept a join request from any User, and private Rooms will only
// accept a join request from someone who is on it's invite list. Only the owner of the Room or the server itself can invite
// Users to a private Room. But remember, just because a User owns a private room doesn't mean the server cannot also invite
// to the room via *Room.AddInvite() function.
//
// Rooms have their own variables which can be accessed and changed anytime. A Room variable can
// be anything compatible with interface{}, so pretty much anything. Room variables should mainly be used
// for things about the room itself that don't change very often (or, for instance, are absolutely needed for a joining
// User).
package rooms

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"sync"
)

// Room represents a room on the server that Users can join and leave. Use rooms.New() to make a new Room.
//
// WARNING: When you use a *User object in your code, DO NOT dereference it. Instead, there are
// many methods for *User for retrieving any information about them you could possibly need.
// Dereferencing them could cause data races in the User fields that get locked by mutexes.
type Room struct {
	name     string
	rType    string
	private  bool
	owner    string
	maxUsers int

	//mux LOCKS ALL FIELDS BELOW
	mux        sync.Mutex
	inviteList []string
	usersMap   map[string]*RoomUser
	vars       map[string]interface{}
}

// RoomUser is the representation of a User in a Room. Note: These
// are not the Users themselves. There are RoomUser methods that provide the same
// functionality as the matching User methods, but if you really need to get a User type from one of these, use
// users.Get() with the RoomUser's Name() function.
type RoomUser struct {
	name    string
	isGuest bool
	dbID    int

	//mux LOCKS ALL FIELDS BELOW
	mux   *sync.Mutex              // Pointer to the User's mux
	conns map[string]*roomUserConn // User's conns info

}

type roomUserConn struct {
	socket *websocket.Conn
	roomIn **Room                  // Pointer to the User's Room pointer
	vars   *map[string]interface{} // Pointer to the User's variables
}

var (
	//THE Rooms AND Rooms MUTEX
	rooms    map[string]*Room = make(map[string]*Room)
	roomsMux sync.Mutex       //LOCKS rooms

	//SERVER SETTINGS
	serverStarted     bool = false
	serverPaused      bool = false
	serverName        string
	deleteRoomOnLeave bool = true
	multiConnect      bool = false
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   MAKE A NEW ROOM   ////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// New adds a new room to the server. This can be called before or after starting the server.
// Parameters:
//
// - name (string): Name of the Room
//
// - rType (string): Room type name (Note: must be a valid RoomType's name)
//
// - isPrivate (bool): Indicates if the room is private or not
//
// - maxUsers (int): Maximum User capacity (Note: 0 means no limit)
//
// - owner (string): The owner of the room. If provided a blank string, will set the owner to the ServerName from ServerSettings
func New(name string, rType string, isPrivate bool, maxUsers int, owner string) (*Room, error) {
	//REJECT INCORRECT INPUT
	if len(name) == 0 {
		return &Room{}, errors.New("rooms.New() requires a name")
	} else if maxUsers < 0 {
		maxUsers = 0
	} else if owner == "" {
		owner = serverName
	}

	var roomType *RoomType
	var ok bool
	if roomType, ok = roomTypes[rType]; !ok {
		return &Room{}, errors.New("Invalid room type")
	}

	//ADD THE ROOM
	roomsMux.Lock()
	if _, ok := rooms[name]; ok {
		roomsMux.Unlock()
		return &Room{}, errors.New("A Room with the name '" + name + "' already exists")
	}
	userMap := make(map[string]*RoomUser)
	roomVars := make(map[string]interface{})
	invList := []string{}
	theRoom := Room{name: name, private: isPrivate, inviteList: invList, usersMap: userMap, maxUsers: maxUsers, vars: roomVars,
		owner: owner, rType: rType}
	rooms[name] = &theRoom
	room := rooms[name]
	roomsMux.Unlock()

	//CALLBACK
	if roomType.HasCreateCallback() {
		roomType.CreateCallback()(room)
	}

	return room, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   DELETE A ROOM   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Delete deletes the Room from the server. Will also send a room leave message to all the Users in the Room that you can
// capture with the client APIs.
func (r *Room) Delete() error {
	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return errors.New("The room '" + r.name + "' does not exist")
	}

	// MAKE LEAVE MESSAGE
	leaveMessage := helpers.MakeClientResponse(helpers.ClientActionLeaveRoom, nil, helpers.NewError("", 0))

	// GO THROUGH ALL Users IN ROOM
	for _, v := range r.usersMap {
		//CHANGE User's room POINTER TO nil & SEND MESSAGES
		v.mux.Lock()
		for key := range v.conns {
			(*v.conns[key]).socket.WriteJSON(leaveMessage)
			*((*v.conns[key]).roomIn) = nil
		}
		v.mux.Unlock()
	}

	r.usersMap = nil
	r.mux.Unlock()

	// DELETE THE ROOM
	roomsMux.Lock()
	delete(rooms, r.name)
	roomsMux.Unlock()

	// CALLBACK
	rType := roomTypes[r.rType]
	if rType.HasDeleteCallback() {
		rType.DeleteCallback()(r)
	}

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET A ROOM   /////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Get finds a Room on the server. If the room does not exit, an error will be returned.
func Get(roomName string) (*Room, error) {
	//REJECT INCORRECT INPUT
	if len(roomName) == 0 {
		return &Room{}, errors.New("rooms.Get() requires a room name")
	}

	var room *Room
	var ok bool

	roomsMux.Lock()
	if room, ok = rooms[roomName]; !ok {
		roomsMux.Unlock()
		return &Room{}, errors.New("The room '" + roomName + "' does not exist")
	}
	roomsMux.Unlock()

	//
	return room, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ADD A USER   /////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// AddUser adds a User to the Room.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want to make a User join a Room, use
// *User.Join() instead. Using this improperly will break server mechanics and cause errors and/or memory leaks.
func (r *Room) AddUser(userName string, dbID int, isGuest bool, socket *websocket.Conn, roomIn **Room, userVars *map[string]interface{},
	userStatus *int, userMux *sync.Mutex, connID string) error {
	// REJECT INCORRECT INPUT
	if len(userName) == 0 {
		return errors.New("*Room.AddUser() requires a user name")
	} else if socket == nil {
		return errors.New("*Room.AddUser() requires a user socket")
	} else if multiConnect && len(connID) == 0 {
		return errors.New("Must provide a connID when MultiConnect is enabled")
	} else if !multiConnect {
		connID = "1"
	}

	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return errors.New("The room '" + r.name + "' does not exist")
	} else if r.maxUsers != 0 && len(r.usersMap) == r.maxUsers {
		r.mux.Unlock()
		return errors.New("The room '" + r.name + "' is full")
	}

	// CHECK IF THE ROOM IS PRIVATE, OWNER JOINS FREELY
	if r.private && userName != r.owner {
		// IF SO, CHECK IF THIS USER IS ON THE INVITE LIST
		if len(r.inviteList) > 0 {
			for i := 0; i < len(r.inviteList); i++ {
				if (r.inviteList)[i] == userName {
					// INVITED User
					break
				}
				if i == len(r.inviteList)-1 {
					r.mux.Unlock()
					return errors.New("User '" + userName + "' is not on the invite list")
				}
			}
		} else {
			r.mux.Unlock()
			return errors.New("User '" + userName + "' is not on the invite list")
		}
	}

	// CHECK IF USER IS ALREADY IN THE ROOM
	var roomUserExists bool = false
	if rUser, ok := r.usersMap[userName]; ok {
		roomUserExists = true
		if !multiConnect {
			r.mux.Unlock()
			return errors.New("User '" + userName + "' is already in room '" + r.name + "'")
		}
		(*rUser.mux).Lock()
		if _, ok := rUser.conns[connID]; ok {
			r.mux.Unlock()
			(*rUser.mux).Unlock()
			return errors.New("User '" + userName + "' is already in room '" + r.name + "'")
		}
		(*rUser.mux).Unlock()
	}
	userList := r.usersMap
	newConn := roomUserConn{socket: socket, roomIn: roomIn, vars: userVars}
	// ADD User TO ROOM
	if roomUserExists {
		(*r.usersMap[userName]).conns[connID] = &newConn
	} else {
		conns := make(map[string]*roomUserConn)
		conns[connID] = &newConn
		newUser := RoomUser{name: userName, isGuest: isGuest, dbID: dbID, mux: userMux, conns: conns}
		r.usersMap[userName] = &newUser
	}
	r.mux.Unlock()

	// CHANGE USER'S ROOM
	(*userMux).Lock()
	*roomIn = r
	(*userMux).Unlock()

	//
	roomType := roomTypes[r.rType]
	if roomType.BroadcastUserEnter() {
		//BROADCAST ENTER TO USERS IN ROOM
		message := make(map[string]interface{})
		message[helpers.ServerActionUserEnter] = make(map[string]interface{})
		message[helpers.ServerActionUserEnter].(map[string]interface{})["u"] = userName
		message[helpers.ServerActionUserEnter].(map[string]interface{})["g"] = isGuest
		for _, u := range userList {
			u.mux.Lock()
			for _, conn := range u.conns {
				(*conn).socket.WriteJSON(message)
			}
			u.mux.Unlock()
		}
	}

	// CALLBACK
	if roomType.HasUserEnterCallback() {
		roomType.UserEnterCallback()(r, userName)
	}

	// SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionJoinRoom, r.Name(), helpers.NewError("", 0))
	socket.WriteJSON(clientResp)

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE A USER   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoveUser removes a User from the room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when removing a User from a Room with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *Room) RemoveUser(userName string, connID string) error {
	//REJECT INCORRECT INPUT
	if len(userName) == 0 {
		return errors.New("*Room.RemoveUser() requires a user name")
	} else if multiConnect && len(connID) == 0 {
		return errors.New("Must provide a connID when MultiConnect is enabled")
	} else if !multiConnect {
		connID = "1"
	}
	//
	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return errors.New("The room '" + r.name + "' does not exist")
	}
	if _, ok := r.usersMap[userName]; !ok {
		r.mux.Unlock()
		return errors.New("User '" + userName + "' is not in room '" + r.name + "'")
	}
	if _, ok := (*r.usersMap[userName]).conns[connID]; !ok {
		r.mux.Unlock()
		return errors.New("Invalid connID")
	}
	uConn := *((*r.usersMap[userName]).conns[connID])
	uMux := (*r.usersMap[userName]).mux
	delete((*r.usersMap[userName]).conns, connID)
	if len((*r.usersMap[userName]).conns) == 0 {
		delete(r.usersMap, userName)
	}
	userList := r.usersMap
	r.mux.Unlock()
	//
	roomType := roomTypes[r.rType]

	//DELETE THE ROOM IF THE OWNER LEFT AND UserRoomControl IS ENABLED
	if deleteRoomOnLeave && userName == r.owner {
		deleteErr := r.Delete()
		if deleteErr != nil {
			return deleteErr
		}
	} else if roomType.BroadcastUserLeave() {
		//CONSTRUCT LEAVE MESSAGE
		message := make(map[string]interface{})
		message[helpers.ServerActionUserLeave] = make(map[string]interface{})
		message[helpers.ServerActionUserLeave].(map[string]interface{})["u"] = userName

		//SEND MESSAGE TO USERS
		for _, u := range userList {
			u.mux.Lock()
			for _, conn := range u.conns {
				conn.socket.WriteJSON(message)
			}
			u.mux.Unlock()
		}
	}

	// CHANGE USER'S ROOM
	uMux.Lock()
	*(uConn.roomIn) = nil
	uMux.Unlock()

	//CALLBACK
	if roomType.HasUserLeaveCallback() {
		roomType.UserLeaveCallback()(r, userName)
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLeaveRoom, nil, helpers.NewError("", 0))
	uConn.socket.WriteJSON(clientResp)

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ADD TO inviteList   //////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// AddInvite adds a User to a private Room's invite list. This is only meant for internal Gopher Game Server mechanics.
// If you want a User to invite someone to a private room, use the *User.Invite() function instead.
//
// NOTE: You can use this function safely, but remember that private rooms are designed to have an "owner",
// and only the owner should be able to send an invite and revoke an invitation for their Rooms. Also, *User.Invite()
// will send an invite message to the invited User that the client API can easily receive. Though if you wish to make
// your own implementations for this, don't hesitate!
func (r *Room) AddInvite(userName string) error {
	if !r.private {
		return errors.New("Room is not private")
	} else if len(userName) == 0 {
		return errors.New("*Room.AddInvite() requires a userName")
	}

	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return errors.New("The room '" + r.name + "' does not exist")
	}
	for i := 0; i < len(r.inviteList); i++ {
		if r.inviteList[i] == userName {
			r.mux.Unlock()
			return errors.New("User '" + userName + "' is already on the invite list")
		}
	}
	r.inviteList = append(r.inviteList, userName)
	r.mux.Unlock()

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE FROM inviteList   /////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoveInvite removes a User from a private Room's invite list.
//
// NOTE: You can use this function safely, but remember that private rooms are designed to have an "owner",
// and only the owner should be able to send an invite and revoke an invitation for their Rooms. But if you find the
// need to break the rules here, by all means do so!
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want a User to remove someone from the room's
// private invite list, use the *User.RevokeInvite() function instead.
func (r *Room) RemoveInvite(userName string) error {
	if !r.private {
		return errors.New("Room is not private")
	} else if len(userName) == 0 {
		return errors.New("*Room.RemoveInvite() requires a userName")
	}

	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return errors.New("The room '" + r.name + "' does not exist")
	}
	for i := 0; i < len(r.inviteList); i++ {
		if r.inviteList[i] == userName {
			r.inviteList = append(r.inviteList[:i], r.inviteList[i+1:]...)
			break
		}
		if i == len(r.inviteList)-1 {
			r.mux.Unlock()
			return errors.New("User '" + userName + "' is not on the invite list")
		}
	}
	r.mux.Unlock()

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET A ROOM's inviteList   ////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// InviteList gets a private Room's invite list.
func (r *Room) InviteList() ([]string, error) {
	r.mux.Lock()
	if r.usersMap == nil {
		r.mux.Unlock()
		return []string{}, errors.New("The room '" + r.name + "' does not exist")
	}
	list := r.inviteList
	r.mux.Unlock()
	//
	return list, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET A Room's usersMap   //////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetUserMap retrieves all the RoomUsers as a map[string]*RoomUser.
func (r *Room) GetUserMap() (map[string]*RoomUser, error) {
	var err error
	var userMap map[string]*RoomUser

	r.mux.Lock()
	if r.usersMap == nil {
		err = errors.New("The room '" + r.name + "' does not exist")
	} else {
		userMap = r.usersMap
	}
	r.mux.Unlock()

	return userMap, err
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Room ATTRIBUTE READERS   /////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Name gets the name of the Room.
func (r *Room) Name() string {
	return r.name
}

// Type gets the type of the Room.
func (r *Room) Type() string {
	return r.rType
}

// IsPrivate returns true of the Room is private.
func (r *Room) IsPrivate() bool {
	return r.private
}

// Owner gets the name of the owner of the room
func (r *Room) Owner() string {
	return r.owner
}

// MaxUsers gets the maximum User capacity of the Room.
func (r *Room) MaxUsers() int {
	return r.maxUsers
}

// NumUsers gets the number of Users in the Room.
func (r *Room) NumUsers() int {
	m, e := r.GetUserMap()
	if e != nil {
		return 0
	}
	return len(m)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   RoomUser ATTRIBUTE READERS   /////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Name gets the name of the RoomUser.
func (u *RoomUser) Name() string {
	return u.name
}

// IsGuest returns true if the RoomUser is a guest.
func (u *RoomUser) IsGuest() bool {
	return u.isGuest
}

// DatabaseID gets the database index of the RoomUser.
func (u *RoomUser) DatabaseID() int {
	return u.dbID
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   ///////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetServerStarted is for Gopher Game Server internal mechanics.
func SetServerStarted(val bool) {
	if !serverStarted {
		serverStarted = val
	}
}

// SettingsSet is for Gopher Game Server internal mechanics.
func SettingsSet(name string, deleteOnLeave bool, multiConn bool) {
	if !serverStarted {
		serverName = name
		deleteRoomOnLeave = deleteOnLeave
		multiConnect = multiConn
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER PAUSE AND RESUME   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause is only for internal Gopher Game Server mechanics.
func Pause() {
	if !serverPaused {
		serverPaused = true
		serverStarted = false
	}
}

// Resume is only for internal Gopher Game Server mechanics.
func Resume() {
	if serverPaused {
		serverStarted = true
		serverPaused = false
	}
}
