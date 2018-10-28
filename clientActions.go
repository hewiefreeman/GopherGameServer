package gopher

import (
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/mssola/user_agent"
	"github.com/gorilla/websocket"
	//"encoding/json"
	"errors"
)

const (
	clientActionLogin = "li"
	clientActionLogout = "lo"
	clientActionJoinRoom = "j"
	clientActionLeaveRoom = "lr"
	clientActionChatMessage = "c"
)

func clientActionHandler(action clientAction, userName *string, conn *websocket.Conn, ua *user_agent.UserAgent) (interface{}, error) {
	switch _action := action.A; _action {
		case "login":
			return clientActionLogin(action.P, userName, conn);
		default:
			return nil, errors.New("Unrecognized client action");
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//   BUILT-IN CLIENT ACTIONS   ///////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

func clientActionLogin(params interface{}, userName *string, conn *websocket.Conn) (interface{}, error) {
	pMap := params.(map[string]interface{});
	user, err := users.Login(pMap["n"].(string), -1, pMap["g"].(bool), conn);
	if(err != nil){ return nil, err }
	*userName = user.Name();
	return user.Name(), nil;
}
