package users

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/rooms"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetVariable sets a User variable.
func (u *User) SetVariable(key string, value interface{}) error {
	//REJECT INCORRECT INPUT
	if len(key) == 0 {
		return errors.New("*User.SetVariable() requires a key")
	} else if u.room == "" {
		return errors.New("User '" + u.name + "' must be in a room to set their variables")
	}

	//GET User's CURRENT ROOM
	room, roomErr := rooms.Get(u.room)
	if roomErr != nil {
		return roomErr
	}

	//SET User's VARIABLE
	addErr := room.SetUserVariable(u.name, key, value)

	//
	return addErr
}

// GetVariable gets one of the User's variables.
func (u *User) GetVariable(key string) interface{} {
	//REJECT INCORRECT INPUT
	if len(key) == 0 || u.room == "" {
		return nil
	}

	//GET User's CURRENT ROOM
	room, roomErr := rooms.Get(u.room)
	if roomErr != nil {
		return nil
	}

	//GET User's VARIABLE
	theVar := room.GetUserVariable(u.name, key)

	//
	return theVar
}
