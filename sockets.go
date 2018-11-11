package gopher

import (
	"github.com/gorilla/websocket"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/mssola/user_agent"
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
		if origin != host && origin != hostAlias {
			http.Error(w, "Origin not allowed.", 403)
			return
		}
	}

	//REJECT IF SERVER IS FULL
	if !conns.add() {
		http.Error(w, "Server is full.", 413)
		return
	}

	//UPGRADE CONNECTION
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not establish a connection.", http.StatusForbidden)
		return
	}

	//GET USER AGENT
	ua := user_agent.New(r.Header.Get("User-Agent"))

	//
	go clientActionListener(conn, ua)
}

func clientActionListener(conn *websocket.Conn, ua *user_agent.UserAgent) {
	// CLIENT ACTION INPUT
	var action clientAction

	// THE CLIENT'S User NAME
	var userName string
	// Room THE CLIENT'S CURRENTLY IN
	var roomIn rooms.Room = rooms.Room{}

	// THE CLIENT'S AUTOLOG INFO
	var deviceTag string = ""
	var devicePass string = ""
	var deviceUserID int = 0

	if (*settings).RememberMe {
		var sentTagRequest bool = false
		//PARAMS
		var ok bool
		var err error
		var pMap map[string]interface{}
		var oldPass string = ""
		//PING-PONG FOR TAGGING DEVICE - BREAKS WHEN THE DEVICE HAS BEEN PROPERLY TAGGED OR AUTHENTICATED.
		for {
			if !sentTagRequest {
				//SEND TAG RETRIEVAL MESSAGE
				tagMessage := make(map[string]interface{})
				tagMessage[helpers.ServerActionRequestDeviceTag] = nil
				writeErr := conn.WriteJSON(tagMessage)
				if writeErr != nil {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				sentTagRequest = true
			}
			//READ INPUT BUFFER
			readErr := conn.ReadJSON(&action)
			if readErr != nil || action.A == "" {
				conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
				conn.Close()
				return
			}

			//DETERMINE ACTION
			if action.A == "0" {
				//NO DEVICE TAG. MAKE ONE AND SEND IT.
				newDeviceTag, newDeviceTagErr := helpers.GenerateSecureString(32)
				if newDeviceTagErr != nil {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				deviceTag = string(newDeviceTag)
				tagMessage := make(map[string]interface{})
				tagMessage[helpers.ServerActionSetDeviceTag] = deviceTag
				writeErr := conn.WriteJSON(tagMessage)
				if writeErr != nil {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
			} else if action.A == "1" {
				//THE CLIENT ONLY HAS A DEVICE TAG, BREAK
				if sentDeviceTag, ohK := action.P.(string); ohK {
					if len(deviceTag) > 0 && sentDeviceTag != deviceTag {
						//CLIENT DIDN'T USE THE PROVIDED DEVICE CODE FROM THE SERVER
						conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
						conn.Close()
						return
					}
					//SEND AUTO-LOG NOT FILED MESSAGE
					notFiledMessage := make(map[string]interface{})
					notFiledMessage[helpers.ServerActionAutoLoginNotFiled] = nil
					writeErr := conn.WriteJSON(notFiledMessage)
					if writeErr != nil {
						conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
						conn.Close()
						return
					}
				} else {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}

				//
				break

			} else if action.A == "2" {
				//THE CLIENT HAS A LOGIN KEY PAIR - MAKE A NEW PASS FOR THEM
				devicePass, err = helpers.GenerateSecureString(32)
				if err != nil {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				//GET PARAMS
				if pMap, ok = action.P.(map[string]interface{}); !ok {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				if deviceTag, ok = pMap["dt"].(string); !ok {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				if oldPass, ok = pMap["da"].(string); !ok {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				var deviceUserIDStr string
				if deviceUserIDStr, ok = pMap["di"].(string); !ok {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				//CONVERT di TO INT
				deviceUserID, err = strconv.Atoi(deviceUserIDStr)
				if err != nil {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				//CHANGE THE CLIENT'S PASS
				newPassMessage := make(map[string]interface{})
				newPassMessage[helpers.ServerActionSetAutoLoginPass] = devicePass
				writeErr := conn.WriteJSON(newPassMessage)
				if writeErr != nil {
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
			} else if action.A == "3" {
				if deviceTag == "" || oldPass == "" || deviceUserID == 0 || devicePass == "" {
					//IRRESPONSIBLE USAGE
					conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
					conn.Close()
					return
				}
				//AUTO-LOG THE CLIENT
				clientName, autoLogErr := users.AutoLogIn(deviceTag, oldPass, devicePass, deviceUserID, conn)
				if autoLogErr != nil {
					//ERROR AUTO-LOGGING - RUN AUTOLOGCOMPLETE AND DELETE KEYS FOR CLIENT, AND SILENTLY CHANGE DEVICE KEY
					newTag, newTagErr := helpers.GenerateSecureString(32)
					if newTagErr != nil {
						conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
						conn.Close()
						return
					}
					autologMessage := make(map[string]interface{})
					autologMessage[helpers.ServerActionAutoLoginFailed] = make(map[string]interface{})
					autologMessage[helpers.ServerActionAutoLoginFailed].(map[string]interface{})["dt"] = newTag
					autologMessage[helpers.ServerActionAutoLoginFailed].(map[string]interface{})["e"] = autoLogErr.Error()
					writeErr := conn.WriteJSON(autologMessage)
					if writeErr != nil {
						conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
						conn.Close()
						return
					}
					deviceTag = newTag
					//
					break
				}
				userName = clientName
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
			sockedDropped(userName)
			conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
			conn.Close()
			return
		}

		//TAKE ACTION
		responseVal, respond, actionErr := clientActionHandler(action, &userName, &roomIn, conn, ua, &deviceTag, &devicePass, &deviceUserID)

		if respond {
			//SEND RESPONSE
			if writeErr := conn.WriteJSON(helpers.MakeClientResponse(action.A, responseVal, actionErr)); writeErr != nil {
				//DISCONNECT USER
				sockedDropped(userName)
				conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second*1))
				conn.Close()
				return
			}
		}

		//
		action = clientAction{}
	}
}

func sockedDropped(userName string) {
	if userName != "" {
		//CLIENT WAS LOGGED IN. LOG THEM OUT
		user, err := users.Get(userName)
		if err == nil {
			user.LogOut()
		}
	}
	conns.subtract()
}

/////////////////////// HELPERS FOR connections

func (c *connections) add() bool {
	(*c).connsMux.Lock()
	defer (*c).connsMux.Unlock()
	//
	if (*settings).MaxConnections != 0 && (*c).conns == (*settings).MaxConnections {
		return false
	} else if (*settings).MaxConnections != 0 {
		(*c).conns++
	}
	//
	return true
}

func (c *connections) subtract() {
	(*c).connsMux.Lock()
	if (*settings).MaxConnections != 0 {
		(*c).conns--
	}
	(*c).connsMux.Unlock()
}
