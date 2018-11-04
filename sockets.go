package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/helpers"
	"net/http"
	"strconv"
	"github.com/mssola/user_agent"
	"github.com/gorilla/websocket"
	"fmt"
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
	// CLIENT ACTION INPUT
	var action clientAction;

	// THE CLIENT'S User NAME
	var userName string;
	// Room THE CLIENT'S CURRENTLY IN
	var roomIn rooms.Room = rooms.Room{};

	for {
		//READ INPUT BUFFER
		readErr := conn.ReadJSON(&action);
		if(readErr != nil || action.A == ""){
			//DISCONNECT USER
			sockedDropped(userName);
			return;
		}

		//TAKE ACTION
		responseVal, respond, actionErr := clientActionHandler(action, &userName, &roomIn, conn, ua);

		if(respond){
			//SEND RESPONSE
			fmt.Println("sending response for:", action.A);
			if writeErr := conn.WriteJSON(helpers.MakeClientResponse(action.A, responseVal, actionErr)); writeErr != nil {
				//DISCONNECT USER
				sockedDropped(userName);
				return;
			}
		}

		//
		action = clientAction{};
	}
}

func sockedDropped(userName string) {
	if(userName != ""){
		//CLIENT WAS LOGGED IN. LOG THEM OUT
		user, err := users.Get(userName);
		if(err == nil){ user.LogOut(); }
	}
}
