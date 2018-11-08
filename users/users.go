// This package contains all the necessary tools to make and work with Users.
package users

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/gorilla/websocket"
)

// The type User represents a client who has logged into the service. A User can
// be a guest, join/leave/create rooms, and call any client action, including your
// custom client actions. If you are not using the built-in authentication, be aware
// that you will need to make sure any client who has not been authenticated by the server
// can't simply log themselves in through the client API.
type User struct {
	name string
	databaseID int
	isGuest bool

	room string

	status int

	socket *websocket.Conn
}

var (
	users map[string]*User = make(map[string]*User)
	usersActionChan *helpers.ActionChannel = helpers.NewActionChannel()
	serverStarted bool = false
	serverName string = ""
	kickOnLogin bool = false
)

// These represent the four statuses a User could be.
const (
	StatusAvailable = iota // User is available
	StatusInGame // User is in a game
	StatusIdle // User is idle
	StatusOffline // User is offline
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG A USER IN   /////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

//Logs a User in to the service.
func Login(userName string, dbID int, isGuest bool, socket *websocket.Conn) (User, error) {
	//REJECT INCORRECT INPUT
	if(len(userName) == 0){
		return User{}, errors.New("users.Login() requires a user name");
	}else if(userName == serverName){
		return User{}, errors.New("The name '"+userName+"' is unavailable");
	}else if(dbID < -1){
		return User{}, errors.New("users.Login() requires a database ID (or -1 for no ID)");
	}else if(socket == nil){
		return User{}, errors.New("users.Login() requires a socket");
	}

	var err error = nil;

	//ALWAYS SET A GUEST'S id TO -1
	databaseID := dbID
	if isGuest { databaseID = -1 }

	response := usersActionChan.Execute(loginUser, []interface{}{userName, databaseID, isGuest, socket});
	if(response[1] != nil){
		if(kickOnLogin){
			DropUser(userName);
			//TRY AGAIN
			response = usersActionChan.Execute(loginUser, []interface{}{userName, databaseID, isGuest, socket});
			if(response[1] != nil){ return User{}, errors.New("Unexpected error while logging in"); }
		}else{
			err = response[1].(error);
		}
	}

	//SUCCESS, SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogin, userName, nil);
	socket.WriteJSON(clientResp);

	//
	return response[0].(User), err;
}

func loginUser(p []interface{}) []interface{} {
	userName, dbID, isGuest, socket := p[0].(string), p[1].(int), p[2].(bool), p[3].(*websocket.Conn);
	var userRef User = User{};
	var err error = nil;

	if _, ok := users[userName]; ok {
		err = errors.New("User '"+userName+"' is already logged in");
	}else{
		newUser := User{name: userName, databaseID: dbID, isGuest: isGuest, socket: socket};
		users[userName] = &newUser;
		userRef = *users[userName];
	}

	return []interface{}{userRef, err};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   LOG A USER OUT   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Logs a User out from the service.
func (u *User) LogOut() {
	//REMOVE USER FROM THEIR ROOM
	if(u.room != ""){
		room, err := rooms.Get(u.room);
		if(err == nil){
			room.RemoveUser(u.name);
		}
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogout, nil, nil);
	u.socket.WriteJSON(clientResp);

	//LOG USER OUT
	usersActionChan.Execute(logUserOut, []interface{}{u.name});
}

func logUserOut(p []interface{}) []interface{} {
	userName := p[0].(string);
	delete(users, userName);
	return []interface{}{};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET A USER   ////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Gets a User by their name.
func Get(userName string) (User, error) {
	var err error = nil;

	//REJECT INCORRECT INPUT
	if(len(userName) == 0){ return User{}, errors.New("users.Get() requires a user name"); }

	response := usersActionChan.Execute(getUser, []interface{}{userName});
	if(response[1] != nil){ err = response[1].(error); }

	//
	return response[0].(User), err;
}

func getUser(p []interface{}) []interface{} {
	userName := p[0].(string);
	var userRef User = User{};
	var err error = nil;

	if user, ok := users[userName]; ok {
		userRef = *user;
	}else{
		err = errors.New("User '"+userName+"' is not logged in");
	}

	return []interface{}{userRef, err};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   MAKE A USER JOIN/LEAVE A ROOM   /////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Makes a User join a Room.
func (u *User) Join(r rooms.Room) error {
	if(u.room == r.Name()){
		return errors.New("User '"+u.name+"' is already in room '"+r.Name()+"'");
	}else if(u.room != ""){
		//LEAVE USER'S CURRENT ROOM
		u.Leave();
	}

	//CHANGE User's ROOM NAME
	response := usersActionChan.Execute(changeUserRoomName, []interface{}{u, r.Name()});
	if(response[0] != nil){ return response[0].(error); }

	//ADD USER TO DESIGNATED ROOM
	addErr := r.AddUser(u.name, u.isGuest, u.socket, response[1].(*string));
	if(addErr != nil){
		//CHANGE User's ROOM NAME BACK
		response = usersActionChan.Execute(changeUserRoomName, []interface{}{u, ""});
		if(response[0] != nil){ return response[0].(error); }
	}

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionJoinRoom, r.Name(), nil);
	u.socket.WriteJSON(clientResp);

	//
	return nil;
}

// Makes a User leave their current room.
func (u *User) Leave() error {
	if(u.room != ""){
		room, roomErr := rooms.Get(u.room);
		if roomErr != nil { return roomErr; }
		//
		removeErr := room.RemoveUser(u.name);
		if(removeErr != nil){ return removeErr; }
	}else{
		return errors.New("User '"+u.name+"' is not in a room.");
	}

	//CHANGE User's ROOM NAME
	response := usersActionChan.Execute(changeUserRoomName, []interface{}{u, ""});
	if(response[0] != nil){ return response[0].(error) }

	//SEND RESPONSE TO CLIENT
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLeaveRoom, nil, nil);
	u.socket.WriteJSON(clientResp);

	return nil;
}

func changeUserRoomName(p []interface{}) []interface{} {
	theUser, roomName := p[0].(*User), p[1].(string);
	var err error = nil;
	var roomIn *string = nil;

	if _, ok := users[(*theUser).name]; ok {
		(*users[(*theUser).name]).room = roomName;
		(*theUser).room = roomName
		roomIn = &(*users[(*theUser).name]).room;
	}else{
		err = errors.New("User '"+theUser.name+"' is not logged in");
	}

	//
	return []interface{}{err, roomIn}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   GET THE STATUS OF A USER   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Gets the current status of a User by their name.
func GetStatus(userName string) int {
	user, err := Get(userName);
	if(err != nil){ return StatusOffline };
	//
	return user.status;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   INVITE TO User's PRIVATE ROOM   /////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Sends an invite to the specified user by name, provided they are online, the Room is private, and this User
// is the owner of the Room.
func (u *User) Invite(userName string, room rooms.Room) error {
	if(len(userName) == 0){
		return errors.New("*User.Invite() requires a userName");
	}else if(!room.IsPrivate()){
		return errors.New("The room '"+room.Name()+"' is not private");
	}else if(room.Owner() != u.name){
		return errors.New("The user '"+u.name+"' is not the owner of the room '"+room.Name()+"'");
	}

	//GET THE USER
	user, userErr := Get(userName);
	if(userErr != nil){ return userErr; }

	//ADD TO INVITE LIST
	addErr := room.AddInvite(userName);
	if(addErr != nil){ return addErr; }

	//MAKE INVITE MESSAGE
	invMessage := make(map[string]interface{});
	invMessage[helpers.ServerActionRoomInvite] = make(map[string]interface{});
	invMessage[helpers.ServerActionRoomInvite].(map[string]interface{})["u"] = u.name;
	invMessage[helpers.ServerActionRoomInvite].(map[string]interface{})["r"] = room.Name();

	//SEND MESSAGE
	user.socket.WriteJSON(invMessage);

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   REVOKE INVITE TO User's PRIVATE ROOM   //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Revokes an invite to the specified user by name, provided they are online, the Room is private, and this User
// is the owner of the Room.
func (u *User) RevokeInvite(userName string, room rooms.Room) error {
	if(len(userName) == 0){
		return errors.New("*User.RevokeInvite() requires a userName");
	}else if(!room.IsPrivate()){
		return errors.New("The room '"+room.Name()+"' is not private");
	}else if(room.Owner() != u.name){
		return errors.New("The user '"+u.name+"' is not the owner of the room '"+room.Name()+"'");
	}

	//REMOVE FROM INVITE LIST
	removeErr := room.RemoveInvite(userName);
	if(removeErr != nil){ return removeErr; }

	//
	return nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   KICK A USER   ///////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Logs a User out by their name. Also used by KickDupOnLogin in ServerSettings.
func DropUser(userName string) error {
	if(len(userName) == 0){
		return errors.New("users.DropUser() requires a user name");
	}
	//
	user, err := Get(userName);
	if(err != nil){
		return err;
	}
	//
	user.LogOut();
	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   User ATTRIBUTE READERS   ////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Gets the name of the User.
func (u *User) Name() string {
	return u.name;
}

// Gets the database table index of the User.
func (u *User) DatabaseID() int {
	return u.databaseID;
}

// Gets the name of the Room that the User is currently in. If you get a blank string, this simply means
// the User is not in a room.
func (u *User) RoomName() string {
	return u.room;
}

// Gets the WebSocket connection of a User.
func (u *User) Socket() *websocket.Conn {
	return u.socket;
}

// Returns true if the User is a guest.
func (u *User) IsGuest() bool {
	return u.isGuest;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// For Gopher Game Server internal mechanics only.
func SetServerStarted(val bool){
	if(!serverStarted){
		serverStarted = val;
	}
}

// For Gopher Game Server internal mechanics only.
func SettingsSet(kickDups bool, name string, deleteOnLeave bool){
	if(!serverStarted){
		kickOnLogin = kickDups;
		serverName = name;
		rooms.SettingsSet(name, deleteOnLeave, usersActionChan);
	}
}
