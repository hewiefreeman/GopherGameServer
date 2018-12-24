package gopher

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"github.com/hewiefreeman/GopherGameServer/users"
	"github.com/hewiefreeman/GopherGameServer/rooms"
)

func macroListener(){
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Server Command: ")
		text, _ := reader.ReadString('\n')

		stop := handleMacro(text[0:len(text)-2])
		if stop {
			return
		}
	}
}

func handleMacro(macro string) bool {
	if macro == "pause" {
		Pause()
	} else if macro == "resume" {
		Resume()
	} else if macro == "shutdown" {
		ShutDown()
		return true
	} else if len(macro) >= 13 && macro[0:11] == "newroomtype" {
		macroNewRoomType(macro)
	} else if len(macro) >= 9 && macro[0:7] == "newroom" {
		macroNewRoom(macro)
	} else if len(macro) >= 6 && macro[0:4] == "kick" {
		macroKick(macro)
	}
	return false
}

// KICK A USER
func macroKick(macro string) {
	userName := macro[5:len(macro)]
	user, userErr := users.Get(userName)
	if userErr != nil {
		fmt.Println(userErr)
		return
	}
	user.Kick()
	fmt.Println("Kicked user '"+userName+"'")
}

func macroNewRoom(macro string) {
	s := strings.Split(macro, " ")
	if len(s) != 5 {
		fmt.Println("newroom expects 4 parameters (name string, rType string, isPrivate bool, maxUsers int)")
		return
	}
	isPrivate := false
	if s[3] == "true" || s[3] == "t" {
		isPrivate = true
	}
	maxUsers, err := strconv.Atoi(s[4])
	if err != nil {
		fmt.Println("maxUsers must be an integer")
		return
	}
	_, roomErr := rooms.New(s[1], s[2], isPrivate, maxUsers, "")
	if roomErr != nil {
		fmt.Println(roomErr)
		return
	}
	fmt.Println("Created room '"+s[1]+"'")
}

func macroNewRoomType(macro string) {
	s := strings.Split(macro, " ")
	if len(s) != 3 {
		fmt.Println("newroomtype expects 2 parameters (name string, serverOnly bool)")
		return
	}
	serverOnly := false
	if s[2] == "true" || s[2] == "t" {
		serverOnly = true
	}
	rooms.NewRoomType(s[1], serverOnly)
	if _, ok := rooms.GetRoomTypes()[s[1]]; !ok {
		fmt.Println("The server must be paused to create a room type")
		return
	}
	fmt.Println("Created room type '"+s[1]+"'")
}
