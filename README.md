# <img src="https://raw.githubusercontent.com/hewiefreeman/GopherGameServer/master/Server%20Gopher.png" width="140" height="140">Gopher Game Server

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/hewiefreeman/GopherGameServer?status.svg)](https://godoc.org/github.com/hewiefreeman/GopherGameServer) <img src="https://img.shields.io/badge/version-v1.0--beta.1-blue.svg"> [![Go Report Card](https://goreportcard.com/badge/github.com/hewiefreeman/GopherGameServer?f=v101)](https://goreportcard.com/report/github.com/hewiefreeman/GopherGameServer)

Gopher Game Server provides a flexible and diverse set of tools that greatly ease developments of any type of online multiplayer game, or real-time application. GGS does all the heavy lifting for you, ensuring you never need to worry about syncronizing or data type conversions.

Moreover, Gopher has a built-in, fully customizable SQL client authentication mechanism that creates and manages users' accounts for you. It even ties in a friending tool, so users can befriend and invite one another to groups, check each other's status, and more. All components are easily configurable and customizable for any specific project's needs.

### :star: Main features

 - Super easy APIs for server, database, and client coding
 - Chat, private messaging, and voice chat
 - Customizable client authentication (\*1)
 - Built-in friending mechanism (\*1)
 - Supports multiple connections on the same User
 - Server saves state on shut-down and restores on reboot (\*2)

> (\*1) A MySQL (or similar SQL) database is required for the authentication/friending feature, but is an optional (like most) feature that can be enabled or disabled to use your own implementations.

> (\*2) When updating and restarting your server, you might need to be able to recover any rooms that were in the middle of a game. This enables you to do so with minimal effort.

### Upcoming features

 - Distributed load balancer and server coordinator
 - Distributed server broadcasts
 - GUI for administrating and monitoring servers
 - Integration with [GopherDB](https://github.com/hewiefreeman/GopherDB) when stable

# :video_game: Client APIs

 - JavaScript: [GopherClientJS](https://github.com/hewiefreeman/GopherClientJS)
 - Java: [GopherClientJava](https://github.com/hewiefreeman/GopherClientJava) ( **In development** )
 
 > If you want to make a client API in an unsupported language and want to know where to start and/or have any questions, feel free to [email me](mailto:dominiquedebergue@gmail.com?subject=[GitHub]%20Gopher%20Game%20Server)!

# :file_folder: Installing
Gopher Game Server requires at least **Go v1.8+** (and **MySQL v5.7+** for the authentication and friending features).

First, install the dependencies:

    go get github.com/gorilla/websocket
    go get github.com/go-sql-driver/mysql
    go get golang.org/x/crypto/bcrypt

Then install the server:

    go get github.com/hewiefreeman/GopherGameServer

# :books: Usage

[:bookmark: Wiki Home](https://github.com/hewiefreeman/GopherGameServer/wiki)

### Table of Contents

1) [**Getting Started**](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started)
   - [Set-Up](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#blue_book-set-up)
   - [Core Server Settings](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#blue_book-core-server-settings)
   - [Server Callbacks](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#blue_book-server-callbacks)
   - [Macro Commands](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#blue_book-macro-commands)
2) [**Rooms**](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms)
   - [Room Types](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-room-types)
   - [Room Broadcasts](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-room-broadcasts)
   - [Room Callbacks](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-room-callbacks)
   - [Creating & Deleting Rooms](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-creating--deleting-rooms)
   - [Room Variables](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-room-variables)
   - [Messaging](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-messaging)
3) [**Users**](https://github.com/hewiefreeman/GopherGameServer/wiki/Users)
   - [Login & Logout](https://github.com/hewiefreeman/GopherGameServer/wiki/Users#blue_book-login-and-logout)
   - [Joining & Leaving Rooms](https://github.com/hewiefreeman/GopherGameServer/wiki/Users#blue_book-joining--leaving-rooms)
   - [User Variables](https://github.com/hewiefreeman/GopherGameServer/wiki/Users#blue_book-user-variables)
   - [Initiating and Revoking Room Invites](https://github.com/hewiefreeman/GopherGameServer/wiki/Users#blue_book-initiating-and-revoking-room-invites)
   - [User Status](https://github.com/hewiefreeman/GopherGameServer/wiki/Users#blue_book-user-status)
   - [Messaging](https://github.com/hewiefreeman/GopherGameServer/wiki/Users#blue_book-messaging)
4) [**Custom Client Actions**](https://github.com/hewiefreeman/GopherGameServer/wiki/Custom-Client-Actions)
   - [Creating a Custom Client Action](https://github.com/hewiefreeman/GopherGameServer/wiki/Custom-Client-Actions#blue_book-creating-a-custom-client-action)
   - [Responding to a Custom Client Action](https://github.com/hewiefreeman/GopherGameServer/wiki/Custom-Client-Actions#blue_book-responding-to-a-custom-client-action)
6) [**Saving & Restoring**](https://github.com/hewiefreeman/GopherGameServer/wiki/Saving-&-Restoring)
   - [Set-Up](https://github.com/hewiefreeman/GopherGameServer/wiki/Saving-&-Restoring#blue_book-set-up)
5) [**SQL Features**](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features)
   - [Set-Up](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features#blue_book-set-up)
   - [Authenticating Clients](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features#blue_book-authenticating-clients)
   - [Custom Account Info](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features#blue_book-custom-account-info)
   - [Customizing Authentication Features](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features#blue_book-customizing-authentication-features)
   - [Auto-Login (Remember Me)](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features#blue_book-auto-login-remember-me)
   - [Friending](https://github.com/hewiefreeman/GopherGameServer/wiki/SQL-Features#blue_book-friending)

# :scroll: Documentation

[Package gopher](https://godoc.org/github.com/hewiefreeman/GopherGameServer) - Main server package for startup and settings

[Package core](https://godoc.org/github.com/hewiefreeman/GopherGameServer/core) - Package for all User and Room functionality

[Package actions](https://godoc.org/github.com/hewiefreeman/GopherGameServer/actions) - Package for making custom client actions

[Package database](https://godoc.org/github.com/hewiefreeman/GopherGameServer/database) - Package for customizing your database

# :milky_way: Contributions
Contributions are open and welcomed! Help is needed for everything from documentation, cleaning up code, performance enhancements, client APIs and more. Don't forget to show your support with a :star:!

If you want to make a client API in an unsupported language and want to know where to start and/or have any questions, feel free to [email me](mailto:dominiquedebergue@gmail.com?subject=[GitHub]%20Gopher%20Game%20Server)!

Please read the following articles before submitting any contributions or filing an Issue:

 - [Contribution Guidlines](https://github.com/hewiefreeman/GopherGameServer/blob/master/CONTRIBUTING.md)
 - [Code of Conduct](https://github.com/hewiefreeman/GopherGameServer/blob/master/CODE_OF_CONDUCT.md)
