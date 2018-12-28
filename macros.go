package gopher

import (
	"bufio"
	"fmt"
	"github.com/hewiefreeman/GopherGameServer/rooms"
	"github.com/hewiefreeman/GopherGameServer/users"
	"os"
	"strconv"
	"strings"
)

func macroListener() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("[Gopher] Command: ")
		text, _ := reader.ReadString('\n')
		stop := handleMacro(text[0 : len(text)-2])
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
	} else if len(macro) >= 12 && macro[0:10] == "deleteroom" {
		macroDeleteRoom(macro)
	} else if len(macro) >= 9 && macro[0:7] == "newroom" {
		macroNewRoom(macro)
	} else if len(macro) >= 6 && macro[0:4] == "kick" {
		macroKick(macro)
	}
	return false
}

// KICK A USER
func macroKick(macro string) {
	userName := macro[5:]
	user, userErr := users.Get(userName)
	if userErr != nil {
		fmt.Println(userErr)
		return
	}
	user.Kick()
	fmt.Println("Kicked user '" + userName + "'")
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
	fmt.Println("Created room '" + s[1] + "'")
}

func macroDeleteRoom(macro string) {
	s := strings.Split(macro, " ")
	if len(s) != 2 {
		fmt.Println("deleteroom expects 1 parameter (name string)")
		return
	}
	room, roomErr := rooms.Get(s[1])
	if roomErr != nil {
		fmt.Println(roomErr)
		return
	}
	deleteErr := room.Delete()
	if deleteErr != nil {
		fmt.Println(deleteErr)
		return
	}
	fmt.Println("Deleted room '" + s[1] + "'")
}
