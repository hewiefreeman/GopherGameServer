# <div style="height:100px;line-height:100px;text-align:center;vertical-align:center;"><img src="https://github.com/hewiefreeman/GopherGameServer/blob/master/Server%20Gopher.png" width="100px" height="100px">Gopher Game Server</div>
Gopher Game Server is a full-featured game server written in Go. Comes with a client API for JavaScript as well (and eventually Java, C++, and C)

-**PROJECT IN DEVELOPMENT**-

Gopher's aim is to provide all the tools necessary to make any type of online game (or any real-time app/chat) a breeze to develop. Gopher will take care of all server-side synchronizing and data type conversions, so you can recieve client actions, set variables, send messages, and much more without having to worry about a thing!

Gopher uses WebSockets and JSON to pass messages between the clients and the server. JSON enabled the server to be designed to let you pass any data type from client to server (or vice versa) without the need to worry about type conversions on either end. WebSockets makes the server as efficient as possible on the network, since the WebSocket protocol is newer and doesn't send nearly as much header and meta data that HTTP and most other protocols require.

Gopher also has a built-in MySQL client authentication mechanism that manages your users' accounts for you. It even ties in a friending tool, so your users can befriend and invite one another to groups, check each other's status, and more. All easily configurable and customizable for your project's needs.

### Main features:

 - Super easy and flexible APIs for server, database, and client coding
 - Chat, private messaging, and voice chat
 - Customizable client authentication (MySQL or similar SQL database required*)
 - Built-in Friending (MySQL or similar SQL database required*)
 - Supports multiple connections using the same login
 - Server saves state on shut-down and restores on reboot
 - Tools provided for administrating server while running

(**\***) A MySQL (or similar SQL) database is required for the authentication/friending feature, but is an optional (like most) feature that can be enabled or disabled to use your own implementations.

# Client APIs

 - JavaScript: [Gopher Client JS](https://github.com/hewiefreeman/GopherClientJS)

The Java, C++, and C (possibly more with some help) client APIs will be made after completing version 1.0 and the JavaScript client API.

# Installing
Gopher Game Server requires at least **Go v1.8+**, and **MySQL v5.7+** for the authentication and friending features.

Installing the server:
     
    go get github.com/hewiefreeman/GopherGameServer
     
Installing the dependencies:

    go get github.com/gorilla/websocket
    go get github.com/mssola/user_agent
    go get github.com/go-sql-driver/mysql
    go get golang.org/x/crypto/bcrypt
     
# Documentation

[Package gopher](https://godoc.org/github.com/hewiefreeman/GopherGameServer) - Main server package for startup and settings

[Package rooms](https://godoc.org/github.com/hewiefreeman/GopherGameServer/rooms) - Package for using the Room, RoomType, and RoomUser types

[Package users](https://godoc.org/github.com/hewiefreeman/GopherGameServer/users) - Package for using the User type

[Package actions](https://godoc.org/github.com/hewiefreeman/GopherGameServer/actions) - Package for making custom client actions

[Package database](https://godoc.org/github.com/hewiefreeman/GopherGameServer/database) - Package for customizing your database

# Usage

(Coming soon...)

# Contributions

**Gopher Game Server will be open for contributions as soon as version 1.0 is finished. At that time, contribution information will be posted here.**
