package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/users"
	//"github.com/hewiefreeman/GopherGameServer/rooms"
	"net/http"
	"strconv"
	"github.com/mssola/user_agent"
	"github.com/gorilla/websocket"
	//"fmt"
)

type clientAction struct {
	A	string // action
	P 	interface{} // parameters
}

func socketInitializer(w http.ResponseWriter, r *http.Request){
	//DECLINE CONNECTIONS COMING FROM OUTSIDE THE ORIGIN SERVER
	if(settings.OriginOnly){
		origin := r.Header.Get("Origin")+":"+strconv.Itoa(settings.Port);
		host := settings.HostName+":"+strconv.Itoa(settings.Port);
		hostAlias := settings.HostAlias+":"+strconv.Itoa(settings.Port);
		if origin != host && origin != hostAlias {
			http.Error(w, "Origin not allowed", 403);
			return;
		}
	}

	//UPGRADE CONNECTION
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not establish a connection.", http.StatusForbidden);
		return;
	}

	//GET USER AGENT
	ua := user_agent.New(r.Header.Get("User-Agent"));

	//
	go clientActionListener(conn, ua);
}

func clientActionListener(conn *websocket.Conn, ua *user_agent.UserAgent) {
	// CLIENT ACTION INPUT/OUTPUT FORMATS
	var action clientAction;
	var response map[string]interface{};

	// The User attached to this client
	var userName string = "";
	//var roomIn rooms.Room;

	//
	var responseVal interface{} = nil;
	var err error = nil;

	for {
		//READ INPUT BUFFER
		err = conn.ReadJSON(&action);
		if(err != nil || action.A == ""){
			//DISCONNECT USER
			sockedDropped(userName);
			return
		}

		//TAKE ACTION
		responseVal, err = clientActionHandler(action, &userName, conn, ua);
		if(err != nil){
			response = make(map[string]interface{});
			response["cr"] = make(map[string]interface{});
			response["cr"].(map[string]interface{})["a"] = action.A;
			response["cr"].(map[string]interface{})["e"] = err.Error();
		}else{
			response = make(map[string]interface{});
			response["cr"] = make(map[string]interface{});
			response["cr"].(map[string]interface{})["a"] = action.A;
			response["cr"].(map[string]interface{})["r"] = responseVal;
		}

		//SEND RESPONSE
		if err = conn.WriteJSON(response); err != nil {
			//DISCONNECT USER
			sockedDropped(userName);
			return
		}

		//
		action = clientAction{};
		response = nil;
		responseVal = nil;
		err = nil;
	}
}

func sockedDropped(userName string) {
	user, err := users.Get(userName);
	if(err == nil){
		//CLIENT WAS LOGGED IN. LOG THEM OUT
		user.LogOut();
	}
}
