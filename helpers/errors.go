package helpers

import ()

// GopherError is used when sending an error message to the client API.
type GopherError struct {
	Message string
	ID      int
}

// Client response message error IDs
const (
	ErrorGopherInvalidAction         = iota + 1001 // Invalid client action
	ErrorGopherIncorrectFormat                     // The client's data is not a map/object
	ErrorGopherIncorrectCustomAction               // Incorrect custom client action type
	ErrorGopherNotLoggedIn                         // The client must be logged in to take action
	ErrorGopherLoggedIn                            // The client must be logged out to take action
	ErrorGopherStatusChange                        // Error while changing User's status
	ErrorGopherFeatureDisabled                     // The feature required for the action is disabled
	ErrorGopherColumnsFormat                       // The client's custom columns data is not a map/object
	ErrorGopherNameFormat                          // The client's user name data is not a string
	ErrorGopherPasswordFormat                      // The client's password data is not a string
	ErrorGopherRememberFormat                      // The client's remember-me data is not a boolean
	ErrorGopherGuestFormat                         // The client's guest data is not a boolean
	ErrorGopherNewPasswordFormat
	ErrorGopherRoomNameFormat
	ErrorGopherRoomTypeFormat
	ErrorGopherPrivateFormat
	ErrorGopherMaxRoomFormat
	ErrorGopherRoomControl
	ErrorGopherServerRoom
	ErrorGopherNotOwner
	ErrorGopherLogin
	ErrorGopherSignUp
	ErrorGopherJoin
	ErrorGopherLeave
	ErrorGopherCreateRoom
	ErrorGopherDeleteRoom
	ErrorGopherInvite
	ErrorGopherRevokeInvite
	ErrorGopherFriendRequest
	ErrorGopherFriendAccept
	ErrorGopherFriendDecline
	ErrorGopherFriendRemove

	// Authentication
	ErrorAuthUnexpected
	ErrorAuthAlreadyLogged
	ErrorAuthRequiredName
	ErrorAuthRequiredPass
	ErrorAuthRequiredNewPass
	ErrorAuthRequiredID
	ErrorAuthRequiredSocket
	ErrorAuthNameUnavail
	ErrorAuthMaliciousChars
	ErrorAuthIncorrectCols
	ErrorAuthInsufficientCols
	ErrorAuthEncryption
	ErrorAuthQuery
	ErrorAuthIncorrectLogin
	ErrorDatabaseInvalidAutolog
	ErrorAuthConversion

	// Misc errors
	ErrorActionDenied
	ErrorServerPaused
)

// NewError creates a new GopherError.
func NewError(message string, id int) GopherError {
	return GopherError{Message: message, ID: id}
}
