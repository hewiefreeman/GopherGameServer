package gopher

import (
	"errors"
	"github.com/hewiefreeman/GopherGameServer/core"
	"github.com/hewiefreeman/GopherGameServer/database"
	"net/http"
)

const (
	// ErrorIncorrectFunction is thrown when function input or return parameters don't match with the callback
	ErrorIncorrectFunction = "Incorrect function parameters or return parameters"
	// ErrorServerRunning is thrown when an action cannot be taken because the server is running. Pausing the server
	// will enable you to run the command.
	ErrorServerRunning = "Cannot call when the server is running."
)

// SetStartCallback sets the callback that triggers when the server first starts up. The
// function passed must have the same parameter types as the following example:
//
//    func serverStarted(){
//	     //code...
//	 }
func SetStartCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func()); ok {
		startCallback = callback
	} else {
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

// SetPauseCallback sets the callback that triggers when the server is paused. The
// function passed must have the same parameter types as the following example:
//
//    func serverPaused(){
//	     //code...
//	 }
func SetPauseCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func()); ok {
		pauseCallback = callback
	} else {
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

// SetResumeCallback sets the callback that triggers when the server is resumed after being paused. The
// function passed must have the same parameter types as the following example:
//
//    func serverResumed(){
//	     //code...
//	 }
func SetResumeCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func()); ok {
		resumeCallback = callback
	} else {
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

// SetShutDownCallback sets the callback that triggers when the server is shut down. The
// function passed must have the same parameter types as the following example:
//
//    func serverStopped(){
//	     //code...
//	 }
func SetShutDownCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func()); ok {
		stopCallback = callback
	} else {
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

// SetClientConnectCallback sets the callback that triggers when a client connects to the server. The
// function passed must have the same parameter types as the following example:
//
//    func clientConnected(writer *http.ResponseWriter, request *http.Request) bool {
//	     //code...
//	 }
//
// The function returns a boolean. If false is returned, the client will receive an HTTP error `http.StatusForbidden` and
// will be rejected from the server. This can be used to, for instance, make a black/white list or implement client sessions.
func SetClientConnectCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(*http.ResponseWriter, *http.Request) bool); ok {
		clientConnectCallback = callback
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}

// SetLoginCallback sets the callback that triggers when a client logs in as a User. The
// function passed must have the same parameter types as the following example:
//
//    func clientLoggedIn(userName string, databaseID int, receivedColumns map[string]interface{}, clientColumns map[string]interface{}) bool {
//	     //code...
//	 }
//
// `userName` is the name of the User logging in, `databaseID` is the index of the User on the database, `receivedColumns` are the custom `AccountInfoColumn` (keys) and their values
// received from the database, and `clientColumns` have the same keys as the `receivedColumns`, but are the input from the client.
//
// The function returns a boolean. If false is returned, the client will receive a `helpers.ErrorActionDenied` (1052) error and will be
// denied from logging in. This can be used to, for instance, suspend or ban a User.
//
// Note: the `clientColumns` decides which `AccountInfoColumn`s were fetched from the database, so the keys will always be the same as `receivedColumns`.
// You can compare the `receivedColumns` and `clientColumns` to, for instance, compare the key 'email' to make sure the
// client also provided the right email address for that account on the database.
func SetLoginCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(string, int, map[string]interface{}, map[string]interface{}) bool); ok {
		if (*settings).EnableSqlFeatures {
			database.LoginCallback = callback
		} else {
			core.LoginCallback = callback
		}
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}

// SetLogoutCallback sets the callback that triggers when a client logs out from a User. The
// function passed must have the same parameter types as the following example:
//
//    func clientLoggedOut(userName string, databaseID int) {
//	     //code...
//	 }
//
// `userName` is the name of the User logging in, `databaseID` is the index of the User on the database.
func SetLogoutCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(string, int)); ok {
		core.LogoutCallback = callback
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}

// SetSignupCallback sets the callback that triggers when a client makes an account. The
// function passed must have the same parameter types as the following example:
//
//    func clientSignedUp(userName string, clientColumns map[string]interface{}) bool {
//	     //code...
//	 }
//
// `userName` is the name of the User logging in, `clientColumns` is the input from the client for setting
// custom `AccountInfoColumn`s on the database.
//
// The function returns a boolean. If false is returned, the client will receive a `helpers.ErrorActionDenied` (1052) error and will be
// denied from signing up. This can be used to, for instance, deny user names or `AccountInfoColumn`s with profanity.
func SetSignupCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(string, map[string]interface{}) bool); ok {
		database.SignUpCallback = callback
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}

// SetDeleteAccountCallback sets the callback that triggers when a client deletes their account. The
// function passed must have the same parameter types as the following example:
//
//    func clientDeletedAccount(userName string, databaseID int, receivedColumns map[string]interface{}, clientColumns map[string]interface{}) bool {
//	     //code...
//	 }
//
// `userName` is the name of the User deleting their account, `databaseID` is the index of the User on the database, `receivedColumns` are the custom `AccountInfoColumn` (keys) and their values
// received from the database, and `clientColumns` have the same keys as the `receivedColumns`, but are the input from the client.
//
// The function returns a boolean. If false is returned, the client will receive a `helpers.ErrorActionDenied` (1052) error and will be
// denied from deleting the account. This can be used to, for instance, make extra input requirements for this action.
//
// Note: the `clientColumns` decides which `AccountInfoColumn`s were fetched from the database, so the keys will always be the same as `receivedColumns`.
// You can compare the `receivedColumns` and `clientColumns` to, for instance, compare the keys named 'email' to make sure the
// client also provided the right email address for that account on the database.
func SetDeleteAccountCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(string, int, map[string]interface{}, map[string]interface{}) bool); ok {
		database.DeleteAccountCallback = callback
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}

// SetAccountInfoChangeCallback sets the callback that triggers when a client changes an `AccountInfoColumn`. The
// function passed must have the same parameter types as the following example:
//
//    func clientChangedAccountInfo(userName string, databaseID int, receivedColumns map[string]interface{}, clientColumns map[string]interface{}) bool {
//	     //code...
//	 }
//
// `userName` is the name of the User changing info, `databaseID` is the index of the User on the database, `receivedColumns` are the custom `AccountInfoColumn` (keys) and their values
// received from the database, and `clientColumns` have the same keys as the `receivedColumns`, but are the input from the client.
//
// The function returns a boolean. If false is returned, the client will receive a `helpers.ErrorActionDenied` (1052) error and will be
// denied from changing the info. This can be used to, for instance, make extra input requirements for this action.
//
// Note: the `clientColumns` decides which `AccountInfoColumn`s were fetched from the database, so the keys will always be the same as `receivedColumns`.
// You can compare the `receivedColumns` and `clientColumns` to, for instance, compare the keys named 'email' to make sure the
// client also provided the right email address for that account on the database.
func SetAccountInfoChangeCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(string, int, map[string]interface{}, map[string]interface{}) bool); ok {
		database.AccountInfoChangeCallback = callback
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}

// SetPasswordChangeCallback sets the callback that triggers when a client changes their password. The
// function passed must have the same parameter types as the following example:
//
//    func clientChangedPassword(userName string, databaseID int, receivedColumns map[string]interface{}, clientColumns map[string]interface{}) bool {
//	     //code...
//	 }
//
// `userName` is the name of the User changing their password, `databaseID` is the index of the User on the database, `receivedColumns` are the custom `AccountInfoColumn` (keys) and their values
// received from the database, and `clientColumns` have the same keys as the `receivedColumns`, but are the input from the client.
//
// The function returns a boolean. If false is returned, the client will receive a `helpers.ErrorActionDenied` (1052) error and will be
// denied from changing the password. This can be used to, for instance, make extra input requirements for this action.
//
// Note: the `clientColumns` decides which `AccountInfoColumn`s were fetched from the database, so the keys will always be the same as `receivedColumns`.
// You can compare the `receivedColumns` and `clientColumns` to, for instance, compare the keys named 'email' to make sure the
// client also provided the right email address for that account on the database.
func SetPasswordChangeCallback(cb interface{}) error {
	if serverStarted {
		return errors.New(ErrorServerRunning)
	} else if callback, ok := cb.(func(string, int, map[string]interface{}, map[string]interface{}) bool); ok {
		database.PasswordChangeCallback = callback
		return nil
	}
	return errors.New(ErrorIncorrectFunction)
}
