# Gopher Game Server
Gopher Game Server is a full-featured game server written in Go. Also comes with a client API for JavaScript (and eventually Java, C++, and C)

Gopher's aim is to provide all the tools necessary to make any type of online game (or any real-time app/chat) a breeze to develop. Gopher will take care of all server-side synchronizing and data type conversions, so you can recieve client actions, set variables, send messages, and much more without having to worry about a thing!

Gopher uses WebSockets and JSON to pass messages between the clients and the server. JSON enabled the server to be designed to let you pass any data type from client to server (or vice versa) without the need to worry about type conversions on either end. WebSockets makes the server as efficient as possible on the network, since the WebSocket protocol is newer and doesn't send nearly as much header and meta data that HTTP and most other protocols require.

-**PROJECT IN DEVELOPMENT**-

**Note**: Gopher Game Server will be open for contributions as soon as version 1.0 is finished.

# Client APIs

 - JavaScript: [Gopher Client JS](https://github.com/hewiefreeman/GopherClientJS)

The Java, C++, and C (possibly more with some help) client APIs will be made after completing version 1.0 and the JavaScript client API.

# Installing
First install the server:
     
    go get github.com/hewiefreeman/GopherGameServer
     
Then install the dependencies:
     
    go get github.com/gorilla/websocket
    go get github.com/mssola/user_agent
     
# Documentation

[Package gopher](https://godoc.org/github.com/hewiefreeman/GopherGameServer) - Main server package for startup and settings

[Package rooms](https://godoc.org/github.com/hewiefreeman/GopherGameServer/rooms) - Package for using the Room type

[Package users](https://godoc.org/github.com/hewiefreeman/GopherGameServer/users) - Package for using the User type

# Usage

(Coming soon...)
