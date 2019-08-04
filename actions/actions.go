// Package actions contains all the tools for making your custom client actions.
package actions

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/core"
	"github.com/hewiefreeman/GopherGameServer/helpers"
)

// CustomClientAction is an action that you can handle on the server from
// a connected client. For instance, a client can send to the server a
// CustomClientAction type "setPosition" that comes with parameters as an object {x: 2, y: 3}.
// You just need to make a callback function for the CustomClientAction type "setPosition", and as soon as the
// action is received by the server, the callback function will be executed concurrently in a Goroutine.
type CustomClientAction struct {
	dataType int

	callback func(interface{}, *Client)
}

// Client objects are created and sent along with your CustomClientAction callback function when a
// client sends an action.
type Client struct {
	action string

	user   *core.User
	connID string
	socket *websocket.Conn

	responded bool
}

// ClientError is used when an error is thrown in your CustomClientAction. Use `actions.NewError()` to make a
// ClientError object. When no error needs to be thrown, use `actions.NoError()` instead.
type ClientError struct {
	message string
	id      int
}

var (
	customClientActions map[string]CustomClientAction = make(map[string]CustomClientAction)

	serverStarted = false
	serverPaused  = false
)

// Default `ClientError`s
const (
	ErrorMismatchedTypes    = iota + 1001 // Client didn't pass the right data type for the given action
	ErrorUnrecognizedAction               // The custom action has not been defined
)

// These are the accepted data types that a client can send with a CustomClientMessage. You must use one
// of these when making a new CustomClientAction, or it will not work. If a client tries to send a type of data that doesnt
// match the type specified for that action, the CustomClientAction will send an error back to the client and skip
// executing your callback function.
const (
	DataTypeBool   = iota // Boolean data type
	DataTypeInt           // int, int32, and int64 data types
	DataTypeFloat         // float32 and float64 data types
	DataTypeString        // string data type
	DataTypeArray         // []interface{} data type
	DataTypeMap           // map[string]interface{} data type
	DataTypeNil           // nil data type
)

// New creates a new `CustomClientAction` with the corresponding parameters:
//
// - actionType (string): The type of action
//
// - (*)callback (func(interface{},Client)): The function that will be executed when a client calls this `actionType`
//
// - dataType (int): The type of data this action accepts. Options are `DataTypeBool`, `DataTypeInt`, `DataTypeFloat`, `DataTypeString`, `DataTypeArray`, `DataTypeMap`, and `DataTypeNil`
//
// (*)Callback function format:
//
//     func yourFunction(actionData interface{}, client *Client) {
//         //...
//
//         // optional client response
//         client.Respond("example", nil);
//     }
//
// - actionData: The data the client sent along with the action
//
// - client: A `Client` object representing the client that sent the action
//
//
// Note: This function can only be called BEFORE starting the server.
func New(actionType string, dataType int, callback func(interface{}, *Client)) error {
	if serverStarted {
		return errors.New("Cannot make a new CustomClientAction once the server has started")
	}
	customClientActions[actionType] = CustomClientAction{
		dataType: dataType,
		callback: callback,
	}
	return nil
}

// NewError creates a new error with a provided message and ID.
func NewError(message string, id int) ClientError {
	return ClientError{message: message, id: id}
}

// NoError is used when no error needs to be thrown in your `CustomClientAction`.
func NoError() ClientError {
	return ClientError{id: -1}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A CustomClientAction RESPONSE TO THE Client   ///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleCustomClientAction handles your custom client actions.
//
// WARNING: This is only meant for internal Gopher Game Server mechanics. Your CustomClientAction callbacks are called
// from this function. This could spawn errors and/or memory leaks.
func HandleCustomClientAction(action string, data interface{}, user *core.User, conn *websocket.Conn, connID string) {
	client := Client{user: user, action: action, socket: conn, connID: connID, responded: false}
	// CHECK IF ACTION EXISTS
	if customAction, ok := customClientActions[action]; ok {
		// CHECK IF THE TYPE OF data MATCHES THE TYPE action SPECIFIES
		if !typesMatch(data, customAction.dataType) {
			client.Respond(nil, NewError("Mismatched data type", ErrorMismatchedTypes))
			return
		}
		//EXECUTE CALLBACK
		customAction.callback(data, &client)
	} else {
		client.Respond(nil, NewError("Unrecognized action", ErrorUnrecognizedAction))
	}
}

//
func typesMatch(data interface{}, theType int) bool {
	switch data.(type) {
	case bool:
		if theType == DataTypeBool {
			return true
		}

	case int:
		if theType == DataTypeInt {
			return true
		}

	case int32:
		if theType == DataTypeInt {
			return true
		}

	case int64:
		if theType == DataTypeInt {
			return true
		}

	case float32:
		if theType == DataTypeFloat {
			return true
		}

	case float64:
		if theType == DataTypeFloat {
			return true
		}

	case string:
		if theType == DataTypeString {
			return true
		}

	case []interface{}:
		if theType == DataTypeArray {
			return true
		}

	case map[string]interface{}:
		if theType == DataTypeMap {
			return true
		}

	case nil:
		if theType == DataTypeNil {
			return true
		}
	}
	//
	return false
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SEND A CustomClientAction RESPONSE TO THE Client   ///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Respond sends a CustomClientAction response to the client. If an error is provided, only the error mesage will be received
// by the Client (the response parameter will not be sent as well). It's perfectly fine to not send back any response if none
// is needed.
//
// NOTE: A response can only be sent once to a Client. Any more calls to Respond() on the same Client will not send a response,
// nor do anything at all. If you want to send a stream of messages to the Client, first get their User object with *Client.User(),
// then you can send data messages directly to the User with the *User.DataMessage() function.
func (c *Client) Respond(response interface{}, err ClientError) {
	//YOU CAN ONLY RESPOND ONCE
	if (*c).responded {
		return
	}
	(*c).responded = true
	//CONSTRUCT MESSAGE
	r := map[string]map[string]interface{}{
		helpers.ServerActionCustomClientActionResponse: {
			"a": (*c).action,
		},
	}
	if err.id != -1 {
		r[helpers.ServerActionCustomClientActionResponse]["e"] = map[string]interface{}{
			"m":  err.message,
			"id": err.id,
		}
	} else {
		r[helpers.ServerActionCustomClientActionResponse]["r"] = response
	}
	//SEND MESSAGE TO CLIENT
	(*c).socket.WriteJSON(r)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Client ATTRIBUTE READERS   ///////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// User gets the *User of the Client.
func (c *Client) User() *core.User {
	return c.user
}

// ConnectionID gets the connection ID of the Client. This is only used if you have MultiConnect enabled in ServerSettings and
// you need to, for instance, call *User functions with the Client's *User obtained with the client.User() function. If
// you do, use client.ConnectionID() when calling any *User functions from a Client.
func (c *Client) ConnectionID() string {
	return c.connID
}

// Action gets the type of action the Client sent.
func (c *Client) Action() string {
	return c.action
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   ///////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SetServerStarted is for Gopher Game Server internal mechanics only.
func SetServerStarted(val bool) {
	if !serverStarted {
		serverStarted = val
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER PAUSE AND RESUME   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause is only for internal Gopher Game Server mechanics.
func Pause() {
	if !serverPaused {
		serverPaused = true
	}
}

// Resume is only for internal Gopher Game Server mechanics.
func Resume() {
	if serverPaused {
		serverPaused = false
	}
}
