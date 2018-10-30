package rooms

import (
	"errors"
)

var (
	roomTypes = make(map[string]RoomType)
)

// A RoomType represents a type of room a client or the server can make.
type RoomType struct {
	serverOnly bool

	broadcastUserEnter bool
	broadcastUserLeave bool

	createCallback func(Room) // roomCreated
	deleteCallback func(Room) // roomDeleted
	userEnterCallback func(Room,string) // roomFrom, userName
	userLeaveCallback func(Room,string) // roomFrom, userName
}

// Adds a RoomType to the server. RoomType is used in conjunction with their cooresponding callbacks.
// Room callbacks are called asynchronously, meaning many Room(s) of a certain type could be executing
// the same callback function at the same time. Though you don't need to worry about this if you are
// only using built-in Gopher server functions (of course ones not labeled with a warning).
//
// A RoomType requires at least the name, serverOnly, and the broadcast parameters. The serverOnly option
// disables clients from being able to manipulate (create/delete/invite) rooms of that type. When broadcastUserEnter is true, a broadcast
// will be sent to all Users in the room notifying them of the entry. You can capture the broadcasts with the client API. Same goes for the
// broadcastUserLeave parameter, but when a User leaves the room. The callbacks can either be set to nil or assigned a function that
// has the same parameter types as shown in the function.
//
// The callbacks work as such:
//  1) createCallback is called when that type of Room is created. The only parameter is a Room that is the created Room.
//  2) deleteCallback is called when that type of Room is deleted. The only parameter is a Room that is the deleted Room.
//  3) userEnterCallback is called when a User enters a Room of that type. The Room parameter is the Room the user entered, and
// the string is the name of the user entering.
//  4) userLeaveCallback is called when a User leaves a Room of that type. The Room parameter is the Room the user left, and
// the string is the name of the user leaving.
//
// Note: This function can only be called BEFORE starting the server.
func NewRoomType(name string, serverOnly bool, broadcastUserEnter bool, broadcastUserLeave bool, createCallback func(Room),
				deleteCallback func(Room), userEnterCallback func(Room,string), userLeaveCallback func(Room,string)) error {
	if(len(name) == 0){
		return errors.New("rooms.RoomType() requires a name");
	}else if(serverStarted){
		return errors.New("Cannot add a Room type once the server has started");
	}
	rt := RoomType{
		serverOnly: serverOnly,

		broadcastUserEnter: broadcastUserEnter,
		broadcastUserLeave: broadcastUserLeave,

		createCallback: createCallback,
		deleteCallback: deleteCallback,
		userEnterCallback: userEnterCallback,
		userLeaveCallback: userLeaveCallback };
	roomTypes[name] = rt;

	//
	return nil;
}

// Gets a slice of all the RoomType names.
func GetRoomTypeNames() []string {
	rTypes := []string{};
	for t, _ := range roomTypes {
		rTypes = append(rTypes, t);
	}
	return rTypes;
}

// Gets a map of all the RoomTypes.
func GetRoomTypes() map[string]RoomType {
	return roomTypes;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   RoomType ATTRIBUTE READERS   ////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *RoomType) ServerOnly() bool {
	return r.serverOnly;
}

func (r *RoomType) BroadcastUserEnter() bool {
	return r.broadcastUserEnter;
}

func (r *RoomType) BroadcastUserLeave() bool {
	return r.broadcastUserLeave;
}

func (r *RoomType) CreateCallback() func(Room) {
	return r.createCallback;
}

func (r *RoomType) DeleteCallback() func(Room) {
	return r.deleteCallback;
}

func (r *RoomType) UserEnterCallback() func(Room,string) {
	return r.userEnterCallback;
}

func (r *RoomType) UserLeaveCallback() func(Room,string) {
	return r.userLeaveCallback;
}
