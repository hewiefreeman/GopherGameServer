package rooms

import (

)

var (
	roomTypes = make(map[string]*RoomType)
)

// A RoomType represents a type of room a client or the server can make. You can only make and set
// options for a RoomType before starting the server. Doing so at any other time will have no effect
// at all.
type RoomType struct {
	serverOnly bool

	voiceChat bool

	broadcastUserEnter bool
	broadcastUserLeave bool

	createCallback func(Room) // roomCreated
	deleteCallback func(Room) // roomDeleted
	userEnterCallback func(Room,string) // roomFrom, userName
	userLeaveCallback func(Room,string) // roomFrom, userName
}

// Adds a RoomType to the server. A RoomType is used in conjunction with it's cooresponding callbacks
// and options. You cannot make a Room on the server until you have at least one RoomType to set it to.
// A RoomType requires at least a name and the serverOnly option, which when set to true will prevent
// the client API from being able to create, destroy, invite or revoke an invitation with that RoomType.
// Though you can always make a CustomClientAction to create a Room, initialize it, send requests, etc.
// When making a new RoomType you can chain the broadcasts and callbacks you want for it like so:
//
//    rooms.NewRoomType("lobby", true).EnableBroadcastUserEnter().EnableBroadcastUserLeave().
//         .SetCreateCallback(yourFunc).SetDeleteCallback(anotherFunc)
//
func NewRoomType(name string, serverOnly bool) *RoomType {
	if(len(name) == 0){
		return &RoomType{};
	}else if(serverStarted){
		return &RoomType{};
	}
	rt := RoomType{
		serverOnly: serverOnly,

		voiceChat: false,

		broadcastUserEnter: false,
		broadcastUserLeave: false,

		createCallback: nil,
		deleteCallback: nil,
		userEnterCallback: nil,
		userLeaveCallback: nil };

	roomTypes[name] = &rt;

	//
	return roomTypes[name];
}

// Gets a map of all the RoomTypes.
func GetRoomTypes() map[string]*RoomType {
	return roomTypes;
}
//////////////////////////////////////////////////////////////////////////////////////////////////////
//   RoomType SETTERS   //////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Enables voice chat for this RoomType.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) EnableVoiceChat() *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).voiceChat = true;
	return r;
}

// When enabled, all Rooms of this RoomType will send an "entry" message to all Users in the Room when another
// User enters the Room. You can capture these messages on the client side easily with the client APIs.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) EnableBroadcastUserEnter() *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).broadcastUserEnter = true;
	return r;
}

// When enabled, all Rooms of this RoomType will send a "left" message to all Users in the Room when another
// User leaves the Room. You can capture these messages on the client side easily with the client APIs.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) EnableBroadcastUserLeave() *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).broadcastUserLeave = true;
	return r;
}

// You can have the server run a function when someone creates a Room of this RoomType by setting the creation
// callback. Your function must take in a Room object as the parameter which is a reference of the created room.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) SetCreateCallback(callback func(Room)) *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).createCallback = callback;
	return r;
}

// You can have the server run a function when someone deletes a Room of this RoomType by setting the delete
// callback. Your function must take in a Room object as the parameter which is a reference of the deleted room.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) SetDeleteCallback(callback func(Room)) *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).deleteCallback = callback;
	return r;
}

// You can have the server run a function when a User enters a Room of this RoomType by setting the User enter callback.
// Your function must take in a Room and a string as the parameters. The Room is the Room in which the User entered,
// and the string is the name of the User that entered.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) SetUserEnterCallback(callback func(Room,string)) *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).userEnterCallback = callback;
	return r;
}

// You can have the server run a function when a User leaves a Room of this RoomType by setting the User leave callback.
// Your function must take in a Room and a string as the parameters. The Room is the Room in which the User left,
// and the string is the name of the User that left.
//
// Note: You must call this BEFORE starting the server in order for it to take effect.
func (r *RoomType) SetUserLeaveCallback(callback func(Room,string)) *RoomType {
	if(serverStarted){
		return r;
	}
	(*r).userLeaveCallback = callback;
	return r;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   RoomType ATTRIBUTE READERS   ////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Returns true if the RoomType can only be manipulated by the server.
func (r *RoomType) ServerOnly() bool {
	return r.serverOnly;
}

// Returns true if voice chat is enabled for this RoomType
func (r *RoomType) VoiceChatEnabled() bool {
	return r.voiceChat;
}

// Returns true if this RoomType has a user entry broadcast
func (r *RoomType) BroadcastUserEnter() bool {
	return r.broadcastUserEnter;
}

// Returns true if this RoomType has a user leave broadcast
func (r *RoomType) BroadcastUserLeave() bool {
	return r.broadcastUserLeave;
}

// Returns the function that this RoomType calls when a Room of this RoomType is created.
func (r *RoomType) CreateCallback() func(Room) {
	return r.createCallback;
}

// Returns true if this RoomType has a creation callback.
func (r *RoomType) HasCreateCallback() bool {
	return r.createCallback != nil;
}

// Returns the function that this RoomType calls when a Room of this RoomType is deleted.
func (r *RoomType) DeleteCallback() func(Room) {
	return r.deleteCallback;
}

// Returns true if this RoomType has a delete callback.
func (r *RoomType) HasDeleteCallback() bool {
	return r.deleteCallback != nil;
}

// Returns the function that this RoomType calls when a User enters a Room of this RoomType.
func (r *RoomType) UserEnterCallback() func(Room,string) {
	return r.userEnterCallback;
}

// Returns true if this RoomType has a user enter callback.
func (r *RoomType) HasUserEnterCallback() bool {
	return r.userEnterCallback != nil;
}

// Returns the function that this RoomType calls when a User leaves a Room of this RoomType.
func (r *RoomType) UserLeaveCallback() func(Room,string) {
	return r.userLeaveCallback;
}

// Returns true if this RoomType has a user leave callback.
func (r *RoomType) HasUserLeaveCallback() bool {
	return r.userLeaveCallback != nil;
}
