package actions

import (

)

// A CustomClientAction is an action that you can handle on the server from
// a connected client. For instance, a client can send to the server a
// custom client action called "setPosition" that comes with an object {x: 2, y: 3}.
// You can have a CustomClientAction that handles this "setPosition", and for instance
// check if that client is in a room, and that position is valid, then set their User variable "position" to the requested
// x/y coordinates and send a data message to all the relevant clients in the room. The possibilities are endless.
type CustomClientAction struct {
	action string
	params interface{}

	userName string

	callback func(string,string,interface{})
}

