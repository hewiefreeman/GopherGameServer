/* This package is used to start and change the core settings for the Gopher Game Server. The
type ServerSettings contains all the parameters for changing the core settings. You can either
pass a ServerSettings when calling Server.Start() or nil if you want to use the default server
settings.*/
package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"math/rand"
	"time"
	"net/http"
	"strconv"
)

// Settings for the Gopher Game Server
type ServerSettings struct {
	HostName string // Server's host name. Use 'https://' for TLS connections. (ex: 'https://example.com')
	HostAlias string // Server's host alias name. Use 'https://' for TLS connections. (ex: 'https://www.example.com')
	IP string // Server's IP address.
	Port int // Server's port.

	TLS bool // Enables TLS/SSL connections.
	CertFile string // SSL/TLS certificate file location (starting from system's root folder).
	PrivKeyFile string // SSL/TLS private key file location (starting from system's root folder).

	OriginOnly bool // When enabled, the server declines connections made from outside the origin server. IMPORTANT: Enable this for web apps and LAN servers.
	KickDupOnLogin bool // When enabled, a logged in User will be disconnected from service when another User logs in with the same name.

	EnableSqlAuth bool // Enables the built-in SQL User authentication. (TO DO)
	SqlIP string // SQL Database IP address. (TO DO)
	SqlPort int // SQL Database port. (TO DO)

	EnableRecovery bool // Enables the recovery of all Rooms, their settings, and their variables on start-up after terminating the server. (TO DO)
	RecoveryLocation string // The folder location (starting from system's root folder) where you would like to store the recovery data. (TO DO)

	EnableAdminTools bool // Enables the use of the Admin Tools (TO DO)
	AdminToolsLogin string // The login name for the Admin Tools (TO DO)
	AdminToolsPassword string // The password for the Admin Tools (TO DO)
}

var (
	settings *ServerSettings
)

// Call with a pointer to your ServerSettings (or nil for defaults) to start the server. The default
// settings are for local testing ONLY. There are security-related attributes in ServerSettings
// for SSL/TLS, connection origin testing, Admin Tools, and more. It's highly recommended to look into
// all ServerSettings attributes.
func Start(s *ServerSettings){
	//SET SERVER SETTINGS
	if(s != nil){
		settings = s;
	}else{
		//DEFAULT localhost SETTINGS
		settings = &ServerSettings{
					HostName: "localhost",
					HostAlias: "localhost",
					IP: "localhost",
					Port: 8080,

					TLS: false,
					CertFile: "",
					PrivKeyFile: "",

					OriginOnly: false,
					KickDupOnLogin: false,

					EnableSqlAuth: false,
					SqlIP: "localhost",
					SqlPort: 3306,

					EnableRecovery: false,
					RecoveryLocation: "C:/",

					EnableAdminTools: false,
					AdminToolsLogin: "admin",
					AdminToolsPassword: "password" }
	}

	//SEED THE rand LIBRARY
	rand.Seed(time.Now().UTC().UnixNano());

	//UPDATE SETTINGS IN PACKAGES
	users.SettingsSet((*settings).KickDupOnLogin);

	//NOTIFY PACKAGES OF SERVER START
	users.SetServerStarted(true);
	rooms.SetServerStarted(true);

	//START HTTP/SOCKET LISTENER
	if(settings.TLS){
		http.HandleFunc("/wss", socketInitializer);
		panic(http.ListenAndServeTLS(settings.IP+":"+strconv.Itoa(settings.Port), settings.CertFile, settings.PrivKeyFile, nil));
	}else{
		http.HandleFunc("/ws", socketInitializer);
		panic(http.ListenAndServe(settings.IP+":"+strconv.Itoa(settings.Port), nil));
	}
}
