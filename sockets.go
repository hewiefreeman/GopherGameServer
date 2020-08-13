package gopher

import (
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/core"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	conns connections = connections{}
)

type connections struct {
	conns    int
	connsMux sync.Mutex
}

type clientAction struct {
	A string      // action
	P interface{} // parameters
}

func socketInitializer(w http.ResponseWriter, r *http.Request) {
	//DECLINE CONNECTIONS COMING FROM OUTSIDE THE ORIGIN SERVER
	if settings.OriginOnly {
		origin := r.Header.Get("Origin") + ":" + strconv.Itoa(settings.Port)
		host := settings.HostName + ":" + strconv.Itoa(settings.Port)
		hostAlias := settings.HostAlias + ":" + strconv.Itoa(settings.Port)
		if origin != host && (settings.HostAlias != "" && origin != hostAlias) {
			http.Error(w, "Origin not allowed.", http.StatusForbidden)
			return
		}
	}

	//REJECT IF SERVER IS FULL
	if !conns.add() {
		http.Error(w, "Server is full.", 413)
		return
	}

	// CLIENT CONNECT CALLBACK
	if clientConnectCallback != nil && !clientConnectCallback(&w, r) {
		http.Error(w, "Could not establish a connection.", http.StatusForbidden)
		return
	}

	//UPGRADE CONNECTION PING-PONG
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not establish a connection.", http.StatusForbidden)
		return
	}

	// START WEBSOCKET LOOP
	go clientActionListener(conn)
}

func clientActionListener(conn *websocket.Conn) {
	// CLIENT ACTION INPUT
	var action clientAction

	var clientMux sync.Mutex // LOCKS user AND connID
	var user *core.User      // THE CLIENT'S User OBJECT
	var connID string        // CLIENT SESSION ID

	// THE CLIENT'S AUTOLOG INFO
	var deviceTag string
	var devicePass string
	var deviceUserID int

	if (*settings).RememberMe {
		//SEND TAG RETRIEVAL MESSAGE
		tagMessage := map[string]interface{}{
			helpers.ServerActionRequestDeviceTag: nil,
		}
		writeErr := conn.WriteJSON(tagMessage)
		if writeErr != nil {
			closeSocket(conn)
			return
		}
		//PARAMS
		var ok bool
		var err error
		var gErr helpers.GopherError
		var oldPass string
		//PING-PONG FOR TAGGING DEVICE - BREAKS WHEN THE DEVICE HAS BEEN PROPERLY TAGGED OR AUTHENTICATED.
		for {
			//READ INPUT BUFFER
			readErr := conn.ReadJSON(&action)
			if readErr != nil || action.A == "" {
				closeSocket(conn)
				return
			}

			//DETERMINE ACTION
			if action.A == "0" {
				//NO DEVICE TAG. MAKE ONE AND SEND IT.
				newDeviceTag, newDeviceTagErr := helpers.GenerateSecureString(32)
				if newDeviceTagErr != nil {
					closeSocket(conn)
					return
				}
				deviceTag = string(newDeviceTag)
				tagMessage := map[string]interface{}{
					helpers.ServerActionSetDeviceTag: deviceTag,
				}
				writeErr := conn.WriteJSON(tagMessage)
				if writeErr != nil {
					closeSocket(conn)
					return
				}
			} else if action.A == "1" {
				//THE CLIENT ONLY HAS A DEVICE TAG, BREAK
				if sentDeviceTag, ohK := action.P.(string); ohK {
					if len(deviceTag) > 0 && sentDeviceTag != deviceTag {
						//CLIENT DIDN'T USE THE PROVIDED DEVICE CODE FROM THE SERVER
						closeSocket(conn)
						return
					}
					//SEND AUTO-LOG NOT FILED MESSAGE
					notFiledMessage := map[string]interface{}{
						helpers.ServerActionAutoLoginNotFiled: nil,
					}
					writeErr := conn.WriteJSON(notFiledMessage)
					if writeErr != nil {
						closeSocket(conn)
						return
					}
				} else {
					closeSocket(conn)
					return
				}

				//
				break

			} else if action.A == "2" {
				//THE CLIENT HAS A LOGIN KEY PAIR - MAKE A NEW PASS FOR THEM
				var pMap map[string]interface{}
				devicePass, err = helpers.GenerateSecureString(32)
				if err != nil {
					closeSocket(conn)
					return
				}
				//GET PARAMS
				if pMap, ok = action.P.(map[string]interface{}); !ok {
					closeSocket(conn)
					return
				}
				if deviceTag, ok = pMap["dt"].(string); !ok {
					closeSocket(conn)
					return
				}
				if oldPass, ok = pMap["da"].(string); !ok {
					closeSocket(conn)
					return
				}
				var deviceUserIDStr string
				if deviceUserIDStr, ok = pMap["di"].(string); !ok {
					closeSocket(conn)
					return
				}
				//CONVERT di TO INT
				deviceUserID, err = strconv.Atoi(deviceUserIDStr)
				if err != nil {
					closeSocket(conn)
					return
				}
				//CHANGE THE CLIENT'S PASS
				newPassMessage := map[string]interface{}{
					helpers.ServerActionSetAutoLoginPass: devicePass,
				}
				writeErr := conn.WriteJSON(newPassMessage)
				if writeErr != nil {
					closeSocket(conn)
					return
				}
			} else if action.A == "3" {
				if deviceTag == "" || oldPass == "" || deviceUserID == 0 || devicePass == "" {
					//IRRESPONSIBLE USAGE
					closeSocket(conn)
					return
				}
				//AUTO-LOG THE CLIENT
				connID, gErr = core.AutoLogIn(deviceTag, oldPass, devicePass, deviceUserID, conn, &user, &clientMux)
				if gErr.ID != 0 {
					//ERROR AUTO-LOGGING - RUN AUTOLOGCOMPLETE AND DELETE KEYS FOR CLIENT, AND SILENTLY CHANGE DEVICE TAG
					newTag, newTagErr := helpers.GenerateSecureString(32)
					if newTagErr != nil {
						closeSocket(conn)
						return
					}
					autologMessage := map[string]map[string]interface{}{
						helpers.ServerActionAutoLoginFailed: {
							"dt": newTag,
							"e": map[string]interface{}{
								"m":  gErr.Message,
								"id": gErr.ID,
							},
						},
					}
					writeErr := conn.WriteJSON(autologMessage)
					if writeErr != nil {
						closeSocket(conn)
						return
					}
					devicePass = ""
					deviceUserID = 0
					deviceTag = newTag
					//
					break
				}
				//
				break
			}
		}
	}

	//STANDARD CONNECTION LOOP
	for {
		//READ INPUT BUFFER
		readErr := conn.ReadJSON(&action)
		if readErr != nil || action.A == "" {
			//DISCONNECT USER
			clientMux.Lock()
			sockedDropped(user, connID, &clientMux)
			closeSocket(conn)
			return
		}

		//TAKE ACTION
		responseVal, respond, actionErr := clientActionHandler(action, &user, conn, &deviceTag, &devicePass, &deviceUserID, &connID, &clientMux)

		if respond {
			//SEND RESPONSE
			if writeErr := conn.WriteJSON(helpers.MakeClientResponse(action.A, responseVal, actionErr)); writeErr != nil {
				//DISCONNECT USER
				clientMux.Lock()
				sockedDropped(user, connID, &clientMux)
				closeSocket(conn)
				return
			}
		}

		//
		action = clientAction{}
	}
}

func closeSocket(conn *websocket.Conn) {
	conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
	conn.Close()
	conns.subtract()
}

func sockedDropped(user *core.User, connID string, clientMux *sync.Mutex) {
	if user != nil {
		//CLIENT WAS LOGGED IN. LOG THEM OUT
		(*clientMux).Unlock()
		user.Logout(connID)
	}
}

/////////////////////// HELPERS FOR connections

func (c *connections) add() bool {
	c.connsMux.Lock()
	//
	if (*settings).MaxConnections != 0 && c.conns == (*settings).MaxConnections {
		c.connsMux.Unlock()
		return false
	}
	c.conns++
	c.connsMux.Unlock()
	//
	return true
}

func (c *connections) subtract() {
	c.connsMux.Lock()
	c.conns--
	c.connsMux.Unlock()
}

// ClientsConnected returns the number of clients connected to the server. Includes connections
// not logged in as a User. To get the number of Users logged in, use the core.UserCount() function.
func ClientsConnected() int {
	conns.connsMux.Lock()
	c := conns.conns
	conns.connsMux.Unlock()
	return c
}
