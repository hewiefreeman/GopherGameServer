// Package gopher is used to start and change the core settings for the Gopher Game Server. The
// type ServerSettings contains all the parameters for changing the core settings. You can either
// pass a ServerSettings when calling Server.Start() or nil if you want to use the default server
// settings.
package gopher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hewiefreeman/GopherGameServer/actions"
	"github.com/hewiefreeman/GopherGameServer/core"
	"github.com/hewiefreeman/GopherGameServer/database"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

/////////// TO DOs:
///////////    - Make authentication for GopherDB
///////////	- Admin tools
///////////    - More useful command-line macros

// ServerSettings are the core settings for the Gopher Game Server. You must fill one of these out to customize
// the server's functionality to your liking.
type ServerSettings struct {
	ServerName     string // The server's name. Used for the server's ownership of private Rooms. (Required)
	MaxConnections int    // The maximum amount of concurrent connections the server will accept. Setting this to 0 means infinite.

	HostName  string // Server's host name. Use 'https://' for TLS connections. (ex: 'https://example.com') (Required)
	HostAlias string // Server's host alias name. Use 'https://' for TLS connections. (ex: 'https://www.example.com')
	IP        string // Server's IP address. (Required)
	Port      int    // Server's port. (Required)

	TLS         bool   // Enables TLS/SSL connections.
	CertFile    string // SSL/TLS certificate file location (starting from system's root folder). (Required for TLS)
	PrivKeyFile string // SSL/TLS private key file location (starting from system's root folder). (Required for TLS)

	OriginOnly bool // When enabled, the server declines connections made from outside the origin server (Admin logins always check origin). IMPORTANT: Enable this for web apps and LAN servers.

	MultiConnect   bool  // Enables multiple connections under the same User. When enabled, will override KickDupOnLogin's functionality.
	MaxUserConns   uint8 // Overrides the default (255) of maximum simultaneous connections on a single User
	KickDupOnLogin bool  // When enabled, a logged in User will be disconnected from service when another User logs in with the same name.

	UserRoomControl   bool // Enables Users to create Rooms, invite/uninvite(AKA revoke) other Users to their owned private rooms, and destroy their owned rooms.
	RoomDeleteOnLeave bool // When enabled, Rooms created by a User will be deleted when the owner leaves. WARNING: If disabled, you must remember to at some point delete the rooms created by Users, or they will pile up endlessly!

	EnableSqlFeatures bool   // Enables the built-in SQL User authentication and friending. NOTE: It is HIGHLY recommended to use TLS over an SSL/HTTPS connection when using the SQL features. Otherwise, sensitive User information can be compromised with network "snooping" (AKA "sniffing").
	SqlIP             string // SQL Database IP address. (Required for SQL features)
	SqlPort           int    // SQL Database port. (Required for SQL features)
	SqlProtocol       string // The protocol to use while comminicating with the MySQL database. Most use either 'udp' or 'tcp'. (Required for SQL features)
	SqlUser           string // SQL user name (Required for SQL features)
	SqlPassword       string // SQL user password (Required for SQL features)
	SqlDatabase       string // SQL database name (Required for SQL features)
	EncryptionCost    int    // The amount of encryption iterations the server will run when storing and checking passwords. The higher the number, the longer encryptions take, but are more secure. Default is 4, range is 4-31.
	CustomLoginColumn string // The custom AccountInfoColumn you wish to use for logging in instead of the default name column.
	RememberMe        bool   // Enables the "Remember Me" login feature. You can read more about this in project's wiki.

	EnableRecovery   bool   // Enables the recovery of all Rooms, their settings, and their variables on start-up after terminating the server.
	RecoveryLocation string // The folder location (starting from system's root folder) where you would like to store the recovery data. (Required for recovery)

	AdminLogin    string // The login name for the Admin Tools (Required for Admin Tools)
	AdminPassword string // The password for the Admin Tools (Required for Admin Tools)
}

type serverRestore struct {
	R map[string]core.RoomRecoveryState
}

var (
	httpServer *http.Server

	settings *ServerSettings

	serverStarted  bool       = false
	serverPaused   bool       = false
	serverStopping bool       = false
	serverEndChan  chan error = make(chan error)

	startCallback         func()
	pauseCallback         func()
	stopCallback          func()
	resumeCallback        func()
	clientConnectCallback func(*http.ResponseWriter, *http.Request) bool

	//SERVER VERSION NUMBER
	version string = "1.0-BETA.2"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Server start-up   ///////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Start will start the server. Call with a pointer to your `ServerSettings` (or nil for defaults) to start the server. The default
// settings are for local testing ONLY. There are security-related options in `ServerSettings`
// for SSL/TLS, connection origin testing, administrator tools, and more. It's highly recommended to look into
// all `ServerSettings` options to tune the server for your desired functionality and security needs.
//
// This function will block the thread that it is ran on until the server either errors, or is manually shut-down. To run code after the
// server starts/stops/pauses/etc, use the provided server callback setter functions.
func Start(s *ServerSettings) {
	if serverStarted || serverPaused {
		return
	}
	serverStarted = true
	fmt.Println("  _______                __\n |   _   |.-----..-----.|  |--..-----..----.\n |.  |___||. _  ||. _  ||.    ||. -__||.  _|\n |.  |   ||:. . ||:. __||: |: ||:    ||: |\n |:  |   |'-----'|: |   '--'--''-----''--'\n |::.. . |       '--' - Game Server -\n '-------'\n\n ")
	fmt.Println("Starting server...")
	// Set server settings
	if s != nil {
		if !s.verify() {
			return
		}
		settings = s
	} else {
		// Default localhost settings
		fmt.Println("Using default settings...")
		settings = &ServerSettings{
			ServerName:     "!server!",
			MaxConnections: 0,

			HostName:  "localhost",
			HostAlias: "localhost",
			IP:        "localhost",
			Port:      8080,

			TLS:         false,
			CertFile:    "",
			PrivKeyFile: "",

			OriginOnly: false,

			MultiConnect:   false,
			KickDupOnLogin: false,

			UserRoomControl:   true,
			RoomDeleteOnLeave: true,

			EnableSqlFeatures: false,
			SqlIP:             "localhost",
			SqlPort:           3306,
			SqlProtocol:       "tcp",
			SqlUser:           "user",
			SqlPassword:       "password",
			SqlDatabase:       "database",
			EncryptionCost:    4,
			CustomLoginColumn: "",
			RememberMe:        false,

			EnableRecovery:   false,
			RecoveryLocation: "C:/",

			AdminLogin:    "admin",
			AdminPassword: "password"}
	}

	// Update package settings
	core.SettingsSet((*settings).KickDupOnLogin, (*settings).ServerName, (*settings).RoomDeleteOnLeave, (*settings).EnableSqlFeatures,
		(*settings).RememberMe, (*settings).MultiConnect, (*settings).MaxUserConns)

	// Notify packages of server start
	core.SetServerStarted(true)
	actions.SetServerStarted(true)
	database.SetServerStarted(true)

	// Start database
	if (*settings).EnableSqlFeatures {
		fmt.Println("Initializing database...")
		dbErr := database.Init((*settings).SqlUser, (*settings).SqlPassword, (*settings).SqlDatabase,
			(*settings).SqlProtocol, (*settings).SqlIP, (*settings).SqlPort, (*settings).EncryptionCost,
			(*settings).RememberMe, (*settings).CustomLoginColumn)
		if dbErr != nil {
			fmt.Println("Database error:", dbErr.Error())
			fmt.Println("Shutting down...")
			return
		}
		fmt.Println("Database initialized")
	}

	// Recover state
	if settings.EnableRecovery {
		recoverState()
	}

	// Start socket listener
	if settings.TLS {
		httpServer = makeServer("/wss", settings.TLS)
	} else {
		httpServer = makeServer("/ws", settings.TLS)
	}

	// Run callback
	if startCallback != nil {
		startCallback()
	}

	// Start macro listener
	go macroListener()

	fmt.Println("Startup complete")

	// Wait for server shutdown
	doneErr := <-serverEndChan

	if doneErr != http.ErrServerClosed {
		fmt.Println("Fatal server error:", doneErr.Error())

		if !serverStopping {
			fmt.Println("Disconnecting users...")

			// Pause server
			core.Pause()
			actions.Pause()
			database.Pause()

			// Save state
			if settings.EnableRecovery {
				saveState()
			}
		}
	}

	fmt.Println("Server shut-down completed")

	if stopCallback != nil {
		stopCallback()
	}
}

func (settings *ServerSettings) verify() bool {
	if settings.ServerName == "" {
		fmt.Println("ServerName in ServerSettings is required. Shutting down...")
		return false

	} else if settings.HostName == "" || settings.IP == "" || settings.Port < 1 {
		fmt.Println("HostName, IP, and Port in ServerSettings are required. Shutting down...")
		return false

	} else if settings.TLS == true && (settings.CertFile == "" || settings.PrivKeyFile == "") {
		fmt.Println("CertFile and PrivKeyFile in ServerSettings are required for a TLS connection. Shutting down...")
		return false

	} else if settings.EnableSqlFeatures == true && (settings.SqlIP == "" || settings.SqlPort < 1 || settings.SqlProtocol == "" ||
		settings.SqlUser == "" || settings.SqlPassword == "" || settings.SqlDatabase == "") {
		fmt.Println("SqlIP, SqlPort, SqlProtocol, SqlUser, SqlPassword, and SqlDatabase in ServerSettings are required for the SQL features. Shutting down...")
		return false

	} else if settings.EnableRecovery == true && settings.RecoveryLocation == "" {
		fmt.Println("RecoveryLocation in ServerSettings is required for server recovery. Shutting down...")
		return false

	} else if settings.EnableRecovery {
		// Check if invalid file location
		if _, err := os.Stat(settings.RecoveryLocation); err != nil {
			fmt.Println("RecoveryLocation error:", err)
			fmt.Println("Shutting down...")
			return false
		}
		var d []byte
		if err := ioutil.WriteFile(settings.RecoveryLocation+"/test.txt", d, 0644); err != nil {
			fmt.Println("RecoveryLocation error:", err)
			fmt.Println("Shutting down...")
			return false
		}
		os.Remove(settings.RecoveryLocation + "/test.txt")

	} else if settings.AdminLogin == "" || settings.AdminPassword == "" {
		fmt.Println("AdminLogin and AdminPassword in ServerSettings are required. Shutting down...")
		return false
	}

	return true
}

func makeServer(handleDir string, tls bool) *http.Server {
	server := &http.Server{Addr: settings.IP + ":" + strconv.Itoa(settings.Port)}
	http.HandleFunc(handleDir, socketInitializer)
	if tls {
		go func() {
			err := server.ListenAndServeTLS(settings.CertFile, settings.PrivKeyFile)
			serverEndChan <- err
		}()
	} else {
		go func() {
			err := server.ListenAndServe()
			serverEndChan <- err
		}()
	}

	//
	return server
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Server actions   ////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Pause will log all Users off and prevent anyone from logging in. All rooms and their variables created by the server will remain in memory.
// Same goes for rooms created by Users unless RoomDeleteOnLeave in ServerSettings is set to true.
func Pause() {
	if !serverPaused {
		serverPaused = true

		fmt.Println("Pausing server...")

		core.Pause()
		actions.Pause()
		database.Pause()

		// Run callback
		if pauseCallback != nil {
			pauseCallback()
		}

		fmt.Println("Server paused")

		serverStarted = false
	}

}

// Resume will allow Users to login again after pausing the server.
func Resume() {
	if serverPaused {
		serverStarted = true

		fmt.Println("Resuming server...")
		core.Resume()
		actions.Resume()
		database.Resume()

		// Run callback
		if resumeCallback != nil {
			resumeCallback()
		}

		fmt.Println("Server resumed")

		serverPaused = false
	}
}

// ShutDown will log all Users off, save the state of the server if EnableRecovery in ServerSettings is set to true, then shut the server down.
func ShutDown() error {
	if !serverStopping {
		serverStopping = true
		fmt.Println("Disconnecting users...")

		// Pause server
		core.Pause()
		actions.Pause()
		database.Pause()

		// Save state
		if settings.EnableRecovery {
			saveState()
		}

		// Shut server down
		fmt.Println("Shutting server down...")
		shutdownErr := httpServer.Shutdown(context.Background())
		if shutdownErr != http.ErrServerClosed {
			return shutdownErr
		}
	}
	//
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   Saving and recovery   ///////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func saveState() {
	fmt.Println("Saving server state...")
	saveErr := writeState(getState(), settings.RecoveryLocation)
	if saveErr != nil {
		fmt.Println("Error saving state:", saveErr)
		return
	}
	fmt.Println("Save state successful")
}

func writeState(stateObj serverRestore, saveFolder string) error {
	state, err := json.Marshal(stateObj)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveFolder+"/Gopher Recovery - "+time.Now().Format("2006-01-02 15-04-05")+".grf", state, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getState() serverRestore {
	return serverRestore{
		R: core.GetRoomsState(),
	}
}

func recoverState() {
	fmt.Println("Recovering previous state...")

	// Get last recovery file
	files, fileErr := ioutil.ReadDir(settings.RecoveryLocation)
	if fileErr != nil {
		fmt.Println("Error recovering state:", fileErr)
		return
	}
	var newestFile string
	var newestTime int64
	for _, f := range files {
		if len(f.Name()) < 19 || f.Name()[0:15] != "Gopher Recovery" {
			continue
		}
		fi, err := os.Stat(settings.RecoveryLocation + "/" + f.Name())
		if err != nil {
			fmt.Println("Error recovering state:", err)
			return
		}
		currTime := fi.ModTime().Unix()
		if currTime > newestTime {
			newestTime = currTime
			newestFile = f.Name()
		}
	}

	// Read file
	r, err := ioutil.ReadFile(settings.RecoveryLocation + "/" + newestFile)
	if err != nil {
		fmt.Println("Error recovering state:", err)
		return
	}

	// Convert JSON
	var recovery serverRestore
	if err = json.Unmarshal(r, &recovery); err != nil {
		fmt.Println("Error recovering state:", err)
		return
	}

	if recovery.R == nil || len(recovery.R) == 0 {
		fmt.Println("No rooms to restore!")
		return
	}

	// Recover rooms
	for name, val := range recovery.R {
		room, roomErr := core.NewRoom(name, val.T, val.P, val.M, val.O)
		if roomErr != nil {
			fmt.Println("Error recovering room '"+name+"':", roomErr)
			continue
		}
		for _, userName := range val.I {
			invErr := room.AddInvite(userName)
			if invErr != nil {
				fmt.Println("Error inviting '"+userName+"' to the room '"+name+"':", invErr)
			}
		}
		room.SetVariables(val.V)
	}

	//
	fmt.Println("State recovery successful")
}
