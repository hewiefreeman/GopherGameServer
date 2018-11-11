// This package contains all the necessary tools to make and work with Users.
package users

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/database"
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

	friends map[string]*database.Friend

	socket *websocket.Conn
}

var (
	users map[string]*User = make(map[string]*User)
	usersActionChan *helpers.ActionChannel = helpers.NewActionChannel()
	serverStarted bool = false
	serverName string = ""
	kickOnLogin bool = false
	sqlFeatures bool = false
	rememberMe bool = false
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

// NOTE: If you are using the SQL authentication features, do not use this! Use the client APIs to log in
// your clients, and you can customize your log in proccess with the database package. Only use this if
// you are making a proper custom authentication for your project.
//
// Logs a User in to the service.
func Login(userName string, dbID int, autologPass string, isGuest bool, remMe bool, socket *websocket.Conn) (User, error) {
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

	//ALWAYS SET A GUEST'S id TO -1
	databaseID := dbID
	if isGuest { databaseID = -1 }

	//GET User's Friend LIST
	var friends []map[string]interface{} = []map[string]interface{}{};
	var friendsMap map[string]*database.Friend;
	var friendsErr error;
	if(dbID != -1 && sqlFeatures){
		friendsMap, friendsErr = database.GetFriends(dbID); // map[string]Friend
		if(friendsErr == nil){
			for _, val := range friendsMap {
				friendEntry := make(map[string]interface{});
				friendEntry["n"] = val.Name();
				friendEntry["rs"] = val.RequestStatus();
				if(val.RequestStatus() == database.FriendStatusAccepted){
					//GET THE User STATUS
					user, userErr := Get(val.Name());
					if(userErr != nil){
						friendEntry["s"] = StatusOffline;
					}else{
						friendEntry["s"] = user.status;
					}
				}
				friends = append(friends, friendEntry);
			}
		}
	}

	response := usersActionChan.Execute(loginUser, []interface{}{userName, databaseID, isGuest, socket, friendsMap});
	if(response[1] != nil){
		if(kickOnLogin){
			DropUser(userName);
			//TRY AGAIN
			response = usersActionChan.Execute(loginUser, []interface{}{userName, databaseID, isGuest, socket, friendsMap});
			if(response[1] != nil){ return User{}, errors.New("Unexpected error while logging in"); }
		}else{
			return User{}, response[1].(error);
		}
	}
	user := response[0].(User);
	//SEND ONLINE MESSAGE TO FRIENDS
	message := make(map[string]interface{});
	message[helpers.ServerActionFriendStatusChange] = make(map[string]interface{});
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = userName;
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = 0;
	for key, val := range user.friends {
		if(val.RequestStatus() == database.FriendStatusAccepted){
			friend, friendErr := Get(key);
			if(friendErr == nil){
				friend.socket.WriteJSON(message);
			}
		}
	}
	//SUCCESS, SEND RESPONSE TO CLIENT
	responseVal := make(map[string]interface{});
	responseVal["n"] = userName;
	responseVal["f"] = friends;
	if(rememberMe && len(autologPass) > 0 && remMe){
		responseVal["ai"] = dbID;
		responseVal["ap"] = autologPass;
	}
	clientResp := helpers.MakeClientResponse(helpers.ClientActionLogin, responseVal, nil);
	socket.WriteJSON(clientResp);

	//
	return user, nil;
}

func loginUser(p []interface{}) []interface{} {
	userName, dbID, isGuest, socket, friends := p[0].(string), p[1].(int), p[2].(bool), p[3].(*websocket.Conn), p[4].(map[string]*database.Friend);
	var userRef User = User{};
	var err error = nil;

	if _, ok := users[userName]; ok {
		err = errors.New("User '"+userName+"' is already logged in");
	}else{
		newUser := User{name: userName, databaseID: dbID, isGuest: isGuest, friends: friends, status: 0, socket: socket};
		users[userName] = &newUser;
		userRef = *users[userName];
	}

	return []interface{}{userRef, err};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   AUTOLOG A USER IN   /////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want the "Remember Me"
// (AKA auto login) feature, enable it in ServerSettings along with the SqlFeatures and cooresponding
// options. You can read more about the "Remember Me" login in the project's usage section.
func AutoLogIn(tag string, pass string, newPass string, dbID int, conn *websocket.Conn) (string, error){
	//VERIFY AND GET USER NAME FROM DATABASE
	userName, autoLogErr := database.AutoLoginClient(tag, pass, newPass, dbID);
	if(autoLogErr != nil){ return "", autoLogErr; }
	//
	_, userErr := Login(userName, dbID, newPass, false, true, conn);
	if(userErr != nil){ return "", userErr; }
	//
	return userName, nil;
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

	//SEND STATUS CHANGE TO FRIENDS
	statusMessage := make(map[string]interface{});
	statusMessage[helpers.ServerActionFriendStatusChange] = make(map[string]interface{});
	statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = u.name;
	statusMessage[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = StatusOffline;
	for key, val := range u.friends {
		if(val.RequestStatus() == database.FriendStatusAccepted){
			friend, friendErr := Get(key);
			if(friendErr == nil){
				friend.socket.WriteJSON(statusMessage);
			}
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
	return []interface{}{err, roomIn};
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SET THE STATUS OF A USER   //////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Sets the status of a User. Also sends a notification to all the User's Friends (with the request
// status "accepted") that they changed their status.
func (u *User) SetStatus(status int) error {
	//CHANGE User's STATUS
	response := usersActionChan.Execute(changeUserStatus, []interface{}{u, status});
	if(response[0] != nil){ return response[0].(error) }

	//SEND STATUS CHANGE MESSAGE TO User's FRIENDS WHOM ARE "ACCEPTED"
	message := make(map[string]interface{});
	message[helpers.ServerActionFriendStatusChange] = make(map[string]interface{});
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["n"] = u.name;
	message[helpers.ServerActionFriendStatusChange].(map[string]interface{})["s"] = status;
	for key, val := range u.friends {
		if(val.RequestStatus() == database.FriendStatusAccepted){
			friend, friendErr := Get(key);
			if(friendErr == nil){
				friend.socket.WriteJSON(message);
			}
		}
	}

	//
	return nil;
}

func changeUserStatus(p []interface{}) []interface{} {
	theUser, theStatus := p[0].(*User), p[1].(int);
	var err error = nil;

	if _, ok := users[(*theUser).name]; ok {
		(*users[(*theUser).name]).status = theStatus;
	}else{
		err = errors.New("User '"+theUser.name+"' is not logged in");
	}

	//
	return []interface{}{err};
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

// Gets the Friend list of the User as a map[string]database.Friend where the key string is the friend's
// User name.
func (u *User) Friends() map[string]*database.Friend {
	return u.friends;
}

// Gets the name of the Room that the User is currently in. If you get a blank string, this simply means
// the User is not in a room.
func (u *User) RoomName() string {
	return u.room;
}

// Gets the status of the User.
func (u *User) Status() int {
	return u.status;
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
func SettingsSet(kickDups bool, name string, deleteOnLeave bool, sqlFeat bool, remMe bool){
	if(!serverStarted){
		kickOnLogin = kickDups;
		serverName = name;
		sqlFeatures = sqlFeat;
		rememberMe = remMe;
		rooms.SettingsSet(name, deleteOnLeave, usersActionChan);
	}
}
