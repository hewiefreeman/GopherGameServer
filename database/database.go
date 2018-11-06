package database

import (
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"errors"
	"strconv"
)

var (
	//DATABASE VARIABLES
	database *sql.DB = nil;
	databaseName string = "";

	//
	serverStarted bool = false
)

//TABLE & COLUMN NAMES
const (
	tableUsers = "users"
	tableFriends = "friends"

	//users TABLE COLUMNS
	usersColumnID = "_id"
	usersColumnName = "name"
	usersColumnPassword = "pass"

	//friends TABLE COLUMNS
	friendsColumnUser = "user"
	friendsColumnFriend = "friend"
	friendsColumnStatus = "status"
)

// WARNING: This is only meant for internal Gopher Game Server mechanics. If you want to enable SQL authorization
// and friending, use the EnableSqlFeatures and cooresponding options in ServerSetting.
func Init(userName string, password string, dbName string, protocol string, ip string, port int) error {
	if(len(userName) == 0){
		 return errors.New("sql.Start() requires a user name");
	}else if(len(password) == 0){
		 return errors.New("sql.Start() requires a password");
	}else if(len(userName) == 0){
		 return errors.New("sql.Start() requires a database name");
	}

	var err error;

	//OPEN THE DATABASE
	var openErr error;
	database, openErr = sql.Open("mysql", userName+":"+password+"@"+protocol+"("+ip+":"+strconv.Itoa(port)+")/"+dbName);
	if(err != nil){ return openErr; }
	//NOTE: Open doesn't open a connection.
	//MUST PING TO CHECK IF FOUND DATABASE
	pingErr := database.Ping();
	if(pingErr != nil){ return errors.New("Could not connect to database!"); }

	databaseName = dbName;

	//CONFIGURE DATABASE
	setupErr := setUp();
	if(setupErr != nil){ return setupErr; }

	//
	return nil;
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   SERVER STARTUP FUNCTIONS   ///////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// For Gopher Game Server internal mechanics.
func SetServerStarted(val bool){
	if(!serverStarted){
		serverStarted = val;
	}
}
