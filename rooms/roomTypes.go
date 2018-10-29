package rooms

import (
	"errors"
)

var (
	roomTypes = make(map[string]RoomType)
)

type RoomType struct {
	name string
	callbacks roomTypeCallbacks
}

type roomTypeCallbacks struct {
	userEnterCallback func(Room,string) //roomFrom, userName
	userLeaveCallback func(Room,string) //roomFrom, userName
}

// Adds a RoomType to the server. RoomType is used in conjunction with their cooresponding callbacks.
// You can assign the callbacks to functions of your own. All Room callbacks are called asynchronously,
// meaning many Room(s) of a certain type could be executing the same callback function at the same time. Though you
// shouldn't need to worry about this if you only use built-in Gopher server functions.
//
// A Room type requires at least a name. The callbacks can either be set to nil or assigned a function that has the same
// parameter types as shown in the rooms.NewRoomType() function.
//
// Note: This function can only be called BEFORE starting the server.
func NewRoomType(name string, userEnterCallback func(Room,string), userLeaveCallback func(Room,string)) error {
	if(len(name) == 0){
		return errors.New("rooms.RoomType() requires a name");
	}else if(serverStarted){
		return errors.New("Cannot add a Room type once the server has started");
	}
	rtc := roomTypeCallbacks{
		userEnterCallback: userEnterCallback,
		userLeaveCallback: userLeaveCallback };
	rt := RoomType{
		name: name,
		callbacks: rtc }
	roomTypes[name] = rt;

	//
	return nil
}

// Get the name of the RoomType.
func (rt *RoomType) Name() string {
	return rt.name;
}
