package helpers

// GopherError is used when sending an error message to the client API.
type GopherError struct {
	Message string
	ID      int
}

// Client response message error IDs
const (
	ErrorGopherInvalidAction         = iota + 1001 // 1001. Invalid client action
	ErrorGopherIncorrectFormat                     // 1002. The client's data is not a map/object
	ErrorGopherIncorrectCustomAction               // 1003. Incorrect custom client action type
	ErrorGopherNotLoggedIn                         // 1004. The client must be logged in to take action
	ErrorGopherLoggedIn                            // 1005. The client must be logged out to take action
	ErrorGopherStatusChange                        // 1006. Error while changing User's status
	ErrorGopherFeatureDisabled                     // 1007. A server feature must be explicitly enabled to take action
	ErrorGopherColumnsFormat                       // 1008. The client's custom columns data is not a map/object
	ErrorGopherNameFormat                          // 1009. The client's user name data is not a string
	ErrorGopherPasswordFormat                      // 1010. The client's password data is not a string
	ErrorGopherRememberFormat                      // 1011. The client's remember-me data is not a boolean
	ErrorGopherGuestFormat                         // 1012. The client's guest data is not a boolean
	ErrorGopherNewPasswordFormat                   // 1013. The client's new password data is not a string
	ErrorGopherRoomNameFormat                      // 1014. The client's room name data is not a string
	ErrorGopherRoomTypeFormat                      // 1015. The client's room type data is not a string
	ErrorGopherPrivateFormat                       // 1016. The client's private room data is not a boolean
	ErrorGopherMaxRoomFormat                       // 1017. The client's maximum room capacity data is not an integer
	ErrorGopherRoomControl                         // 1018. Clients do not have the ability to control rooms
	ErrorGopherServerRoom                          // 1019. The room type specified can only be made by the server
	ErrorGopherNotOwner                            // 1020. The client must be the owner of the room to take action
	ErrorGopherLogin                               // 1021. There was an error logging in
	ErrorGopherSignUp                              // 1022. There was an error signing up
	ErrorGopherJoin                                // 1023. There was an error joining a room
	ErrorGopherLeave                               // 1024. There was an error leaving a room
	ErrorGopherCreateRoom                          // 1025. There was an error creating a room
	ErrorGopherDeleteRoom                          // 1026. There was an error deleting a room
	ErrorGopherInvite                              // 1027. There was an error inviting User to a room
	ErrorGopherRevokeInvite                        // 1028. There was an error revoking a User's invitation to a room
	ErrorGopherFriendRequest                       // 1029. There was an error sending a friend request
	ErrorGopherFriendAccept                        // 1030. There was an error accepting a friend request
	ErrorGopherFriendDecline                       // 1031. There was an error declining a friend request
	ErrorGopherFriendRemove                        // 1032. There was an error removing a friend

	// Authentication
	ErrorAuthUnexpected         // 1033. There was an unexpected authorization error
	ErrorAuthAlreadyLogged      // 1034. The client is already logged in
	ErrorAuthRequiredName       // 1035. A user name is required
	ErrorAuthRequiredPass       // 1036. A password is required
	ErrorAuthRequiredNewPass    // 1037. A new password is required
	ErrorAuthRequiredID         // 1038. An account id is required
	ErrorAuthRequiredSocket     // 1039. A client socket pointer is required
	ErrorAuthNameUnavail        // 1040. The user name is unavailable
	ErrorAuthMaliciousChars     // 1041. There are malicious characters in the client's request variables
	ErrorAuthIncorrectCols      // 1042. The client supplied incorrect custom account info column data
	ErrorAuthInsufficientCols   // 1043. The client supplied an insufficient amount of custom account info columns
	ErrorAuthEncryption         // 1044. There was an error while encrypting data
	ErrorAuthQuery              // 1045. There was an error while querying the database
	ErrorAuthIncorrectLogin     // 1046. The client supplied an incorrect login or password
	ErrorDatabaseInvalidAutolog // 1047. The client supplied incorrect auto-login (remember me) data
	ErrorAuthConversion         // 1048. There was an error while converting data to be stored on the database

	// Misc errors
	ErrorActionDenied // 1049. A callback has denied the server action
	ErrorServerPaused // 1050. The server is paused
)

// NewError creates a new GopherError.
func NewError(message string, id int) GopherError {
	return GopherError{Message: message, ID: id}
}

// NoError creates a new GopherError that represents a state in which no error occurred.
func NoError() GopherError {
	return GopherError{}
}
