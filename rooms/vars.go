package rooms

import (
	"errors"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ROOM VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Set a Room variable.
func (r *Room) SetVariable(key string, value interface{}) error {
	//REJECT INCORRECT INPUT
	if(len(key) == 0){ return errors.New("*Room.SetVariable() requires a key"); }

	response := r.roomVarsActionChannel.Execute(roomVarSet, []interface{}{key, value, r});
	if(len(response) == 0){ return errors.New("Room '"+r.name+"' does not exist"); }

	//
	return nil;
}

func roomVarSet(p []interface{}) []interface{} {
	key, value, room := p[0].(string), p[1], p[2].(*Room)

	(*room.vars)[key] = value;

	//
	return []interface{}{nil}
}

// Get one of the Room's variables.
func (r *Room) GetVariable(key string) (interface{}, error) {
	//REJECT INCORRECT INPUT
	if(len(key) == 0){ return nil, errors.New("*Room.GetVariable() requires a key"); }

	var val interface{} = nil;
	var err error = nil;

	response := r.roomVarsActionChannel.Execute(roomVarGet, []interface{}{key, r});
	if(len(response) == 0){
		err = errors.New("Room '"+r.name+"' does not exist");
	}else{
		val = response[0]
	}
	//
	return val, err;
}

func roomVarGet(p []interface{}) []interface{} {
	key, room := p[0].(string), p[1].(*Room)
	//
	return []interface{}{(*room.vars)[key]}
}

// Get a Map of all the Room variables.
func (r *Room) GetVariableMap() (map[string]interface{}, error) {
	var val map[string]interface{} = nil;
	var err error = nil;

	response := r.roomVarsActionChannel.Execute(roomVarMapGet, []interface{}{r});
	if(len(response) == 0){
		err = errors.New("Room '"+r.name+"' does not exist");
	}else{
		val = response[0].(map[string]interface{})
	}
	//
	return val, err;
}

func roomVarMapGet(p []interface{}) []interface{} {
	room := p[0].(*Room);
	return []interface{}{*room.vars}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   USER VARIABLES   /////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Set a User's variable using their name.
func (r *Room) SetUserVariable(userName string, key string, value interface{}) error {
	//REJECT INCORRECT INPUT
	if(len(key) == 0){
		return errors.New("*Room.SetUserVariable() requires a key");
	}else if(len(userName) == 0){
		return errors.New("*Room.SetUserVariable() requires a user name");
	}

	var err error = nil;

	response := r.usersActionChannel.Execute(userVarSet, []interface{}{userName, key, value, r});
	if(len(response) == 0){
		err = errors.New("Room '"+r.name+"' does not exist");
	}else if(response[0] != nil){
		err = response[0].(error);
	}
	//
	return err;
}

func userVarSet(p []interface{}) []interface{} {
	userName, key, value, room := p[0].(string), p[1].(string), p[2], p[3].(*Room);
	var err error = nil;

	if _, ok := (*room.usersMap)[userName]; ok {
		(*room.usersMap)[userName].vars[key] = value;
	}else{
		err = errors.New("User '"+userName+"' is not in room '"+room.name+"'");
	}

	//
	return []interface{}{err}
}

// Get a User's variable using their name.
func (r *Room) GetUserVariable(userName string, key string) interface{} {
	//REJECT INCORRECT INPUT
	if(len(key) == 0){ return errors.New("*Room.GetUserVariable() requires a key"); }

	var value interface{} = nil;

	response := r.usersActionChannel.Execute(userVarGet, []interface{}{userName, key, r});
	if(len(response) != 0 && response[0] != nil){ value = response[0]; }

	//
	return value;
}

func userVarGet(p []interface{}) []interface{} {
	userName, key, room := p[0].(string), p[1].(string), p[2].(*Room);
	var value interface{} = nil;

	if _, ok := (*room.usersMap)[userName]; ok { value = (*room.usersMap)[userName].vars[key]; }

	//
	return []interface{}{value}
}
