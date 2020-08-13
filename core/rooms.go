package core

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"sync"
)

// Room represents a room on the server that Users can join and leave. Use core.NewRoom() to make a new Room.
//
// WARNING: When you use a *Room object in your code, DO NOT dereference it. Instead, there are
// many methods for *Room for maniupulating and retrieving any information about them you could possibly need.
// Dereferencing them could cause data races in the Room fields that get locked by mutexes.
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

// RoomUser represents a User inside of a Room. Use the *RoomUser.User() function to get a *User from a *RoomUser
type RoomUser struct {
	user *User

	mux   sync.Mutex
	conns map[string]*userConn
}

var (
	rooms    map[string]*Room = make(map[string]*Room)
	roomsMux sync.Mutex
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   MAKE A NEW ROOM   ////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewRoom adds a new room to the server. This can be called before or after starting the server.
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
func NewRoom(name string, rType string, isPrivate bool, maxUsers int, owner string) (*Room, error) {
	//REJECT INCORRECT INPUT
	if len(name) == 0 {
		return &Room{}, errors.New("core.NewRoom() requires a name")
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
	theRoom := Room{name: name, private: isPrivate, inviteList: []string{}, usersMap: make(map[string]*RoomUser), maxUsers: maxUsers,
		vars: make(map[string]interface{}), owner: owner, rType: rType}
	rooms[name] = &theRoom
	roomsMux.Unlock()

	//CALLBACK
	if roomType.HasCreateCallback() {
		roomType.CreateCallback()(&theRoom)
	}

	return &theRoom, nil
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
	leaveMessage := helpers.MakeClientResponse(helpers.ClientActionLeaveRoom, nil, helpers.NoError())

	// GO THROUGH ALL Users IN ROOM
	for _, u := range r.usersMap {
		//CHANGE User's room POINTER TO nil & SEND MESSAGES
		u.mux.Lock()
		for key := range u.conns {
			(*u.conns[key]).socket.WriteJSON(leaveMessage)
			u.user.mux.Lock()
			(*u.conns[key]).room = nil
			u.user.mux.Unlock()
		}
		u.mux.Unlock()
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

// GetRoom finds a Room on the server. If the room does not exit, an error will be returned.
func GetRoom(roomName string) (*Room, error) {
	//REJECT INCORRECT INPUT
	if len(roomName) == 0 {
		return &Room{}, errors.New("core.GetRoom() requires a room name")
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

// AddUser adds a User to the Room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when adding a User to a Room with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *Room) AddUser(user *User, connID string) error {
	userName := user.Name()
	// REJECT INCORRECT INPUT
	if user == nil {
		return errors.New("*Room.AddUser() requires a valid User")
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
		// IF PRIVATE AND NOT OWNER, CHECK IF THIS USER IS ON THE INVITE LIST
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
	var ru *RoomUser
	var ok bool
	if ru, ok = r.usersMap[userName]; ok {
		if !multiConnect {
			r.mux.Unlock()
			return errors.New("User '" + userName + "' is already in room '" + r.name + "'")
		}
		ru.mux.Lock()
		if _, ok := ru.conns[connID]; ok {
			r.mux.Unlock()
			ru.mux.Unlock()
			return errors.New("User '" + userName + "' is already in room '" + r.name + "'")
		}
		ru.mux.Unlock()
	}
	// ADD User TO ROOM
	user.mux.Lock()
	c := user.conns[connID]
	if c == nil {
		r.mux.Unlock()
		return errors.New("Invalid connection ID")
	}
	if ru != nil {
		(*r.usersMap[userName]).mux.Lock()
		(*r.usersMap[userName]).conns[connID] = c
		(*r.usersMap[userName]).mux.Unlock()
	} else {
		conns := make(map[string]*userConn)
		conns[connID] = c
		newUser := RoomUser{user: user, conns: conns}
		r.usersMap[userName] = &newUser
		ru = r.usersMap[userName]
	}
	// CHANGE USER'S ROOM
	c.room = r

	user.mux.Unlock()
	r.mux.Unlock()

	//
	roomType := roomTypes[r.rType]
	if roomType.BroadcastUserEnter() {
		//BROADCAST ENTER TO USERS IN ROOM
		message := map[string]map[string]interface{}{
			helpers.ServerActionUserEnter: {
				"u": userName,
				"g": user.isGuest,
			},
		}
		for _, u := range r.usersMap {
			u.mux.Lock()
			if u.user.Name() != userName {
				for _, conn := range u.conns {
					(*conn).socket.WriteJSON(message)
				}
			}
			u.mux.Unlock()
		}
	}
	// CALLBACK
	if roomType.HasUserEnterCallback() {
		roomType.UserEnterCallback()(r, ru)
	}

	// SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionJoinRoom, r.Name(), helpers.NoError())
	c.socket.WriteJSON(clientResp)

	//
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   REMOVE A USER   //////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoveUser removes a User from the room. If you are using MultiConnect in ServerSettings, the connID
// parameter is the connection ID associated with one of the connections attached to that User. This must
// be provided when removing a User from a Room with MultiConnect enabled. Otherwise, an empty string can be used.
func (r *Room) RemoveUser(user *User, connID string) error {
	//REJECT INCORRECT INPUT
	if user == nil || len(user.name) == 0 {
		return errors.New("*Room.RemoveUser() requires a valid *User")
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
	var ok bool
	var ru *RoomUser
	if ru, ok = r.usersMap[user.name]; !ok {
		r.mux.Unlock()
		return errors.New("User '" + user.name + "' is not in room '" + r.name + "'")
	}
	ru.mux.Lock()
	var uConn *userConn
	if uConn, ok = ru.conns[connID]; !ok {
		r.mux.Unlock()
		ru.mux.Unlock()
		return errors.New("Invalid connID")
	}
	delete(ru.conns, connID)
	// Remove user when no conns are left in room
	if len(ru.conns) == 0 {
		delete(r.usersMap, user.name)
	}
	ru.mux.Unlock()
	userList := r.usersMap
	r.mux.Unlock()
	//
	roomType := roomTypes[r.rType]

	//DELETE THE ROOM IF THE OWNER LEFT AND UserRoomControl IS ENABLED
	if deleteRoomOnLeave && user.name == r.owner {
		deleteErr := r.Delete()
		if deleteErr != nil {
			return deleteErr
		}
	} else if roomType.BroadcastUserLeave() {
		//CONSTRUCT LEAVE MESSAGE
		message := map[string]map[string]interface{}{
			helpers.ServerActionUserLeave: {
				"u": user.name,
			},
		}

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
	user.mux.Lock()
	uConn.room = nil
	user.mux.Unlock()

	//CALLBACK
	if roomType.HasUserLeaveCallback() {
		roomType.UserLeaveCallback()(r, ru)
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLeaveRoom, r.Name(), helpers.NoError())
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
// NOTE: Remember that private rooms are designed to have an "owner",
// and only the owner should be able to send an invite and revoke an invitation for their Rooms. Also, *User.Invite()
// will send an invite notification message to the invited User that the client API can easily receive. Though if you wish to make
// your own implementations for sending and receiving these notifications, this function is safe to use.
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

// RemoveInvite removes a User from a private Room's invite list. To make a User remove someone from their room themselves,
// use the *User.RevokeInvite() function.
//
// NOTE: You can use this function safely, but remember that private rooms are designed to have an "owner",
// and only the owner should be able to send an invite and revoke an invitation for their Rooms. But if you find the
// need to break the rules here, by all means do so!
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
			r.inviteList[i] = r.inviteList[len(r.inviteList)-1]
			r.inviteList = r.inviteList[:len(r.inviteList)-1]
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

// RoomCount returns the number of Rooms created on the server.
func RoomCount() int {
	roomsMux.Lock()
	length := len(rooms)
	roomsMux.Unlock()
	return length
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   RoomUser ATTRIBUTE READERS   /////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// User gets the *User object of a *RoomUser.
func (u *RoomUser) User() *User {
	return u.user
}

// ConnectionIDs returns a []string of all the RoomUser's connection IDs. With MultiConnect in ServerSettings enabled,
// this will give you all the connections for this User that are currently in the Room. Otherwise, if you want
// all the User's connection IDs (not just the connections in the specified Room), use *User.ConnectionIDs() after getting
// the *User object with the *RoomUser.User() function.
func (u *RoomUser) ConnectionIDs() []string {
	u.mux.Lock()
	ids := make([]string, 0, len(u.conns))
	for id := range u.conns {
		ids = append(ids, id)
	}
	u.mux.Unlock()
	return ids
}
