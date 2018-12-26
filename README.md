# <img src="https://raw.githubusercontent.com/hewiefreeman/GopherGameServer/master/Server%20Gopher.png" width="140" height="140">Gopher Game Server

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/hewiefreeman/GopherGameServer?status.svg)](https://godoc.org/github.com/hewiefreeman/GopherGameServer) <img src="https://img.shields.io/badge/version-v1.0--beta.1-blue.svg"> [![Go Report Card](https://goreportcard.com/badge/github.com/hewiefreeman/GopherGameServer)](https://goreportcard.com/report/github.com/hewiefreeman/GopherGameServer)

Gopher Game Server is designed to provide all necessary tools to greatly ease developments of any type of online game (or any real-time app/chat). Gopher will handle all server-side synchronizing and data type conversions, therefore, client actions receiving, variable setting, message sending, and other functionalities are unproblematic.

Moreover, Gopher has a built-in, fully customizable SQL client authentication mechanism that manages users' accounts for you. It even ties in a friending tool, so users can befriend and invite one another to groups, check each other's status, and more. All components are easily configurable and customizable for specific project's needs.

### Main features:

 - Super easy and flexible APIs for server, database, and client coding
 - Chat, private messaging, and voice chat
 - Customizable client authentication (**\***)
 - Built-in Friending (**\***)
 - Supports multiple connections on the same User
 - Server saves state on shut-down and restores on reboot
 - Tools provided for administrating server while running

(**\***) A MySQL (or similar SQL) database is required for the authentication/friending feature, but is an optional (like most) feature that can be enabled or disabled to use your own implementations.

# Client APIs

 - JavaScript: [Gopher Client JS](https://github.com/hewiefreeman/GopherClientJS)

The Java, C++, and C (possibly more with some help) client APIs will be made after completing version 1.0 and the JavaScript client API.

# Installing
Gopher Game Server requires at least **Go v1.8+**, and **MySQL v5.7+** for the authentication and friending features.

First, install the dependencies:

    go get github.com/gorilla/websocket
    go get github.com/go-sql-driver/mysql
    go get golang.org/x/crypto/bcrypt

Then install the server:

    go get github.com/hewiefreeman/GopherGameServer

# Documentation

[Package gopher](https://godoc.org/github.com/hewiefreeman/GopherGameServer) - Main server package for startup and settings

[Package rooms](https://godoc.org/github.com/hewiefreeman/GopherGameServer/rooms) - Package for using the Room, RoomType, and RoomUser types

[Package users](https://godoc.org/github.com/hewiefreeman/GopherGameServer/users) - Package for using the User type

[Package actions](https://godoc.org/github.com/hewiefreeman/GopherGameServer/actions) - Package for making custom client actions

[Package database](https://godoc.org/github.com/hewiefreeman/GopherGameServer/database) - Package for customizing your database

# Usage

**-UNDER CONSTRUCTION-**

### Table of Contents

1) [**Getting Started**](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started)
   - [Set-Up](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#set-up)
   - [Core Server Settings](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#core-server-settings)
   - [Server Callbacks](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#server-callbacks)
   - [Macro Commands](https://github.com/hewiefreeman/GopherGameServer/wiki/Getting-Started#macro-commands)
2) [**Rooms**](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms)
   - [Room Types](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#room-types)
   - [Room Broadcasts](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#room-broadcasts)
   - [Room Callbacks](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#room-callbacks)
   - [Creating & Deleting Rooms](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#creating--deleting-rooms)
   - [Room Variables](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#room-variables)
   - [Messaging](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#messaging)
3) **Users**
   - Login & Logout
   - Joining & Leaving Rooms
   - User Variables
   - Initiating and Revoking Room Invites
   - Messaging
4) **Custom Client Actions**
   - Creating a Custom Client Action
   - Responding to a Custom Client Action
6) **Saving & Restoring**
   - Set-Up
5) **MySQL Features**
   - Set-Up
   - Securely Authenticating Clients
   - Customize Authentication Features
   - Custom Account Info
   - Friending

# Contributions
Contributions are open and welcomed! Help is needed for everything from documentation, cleaning up code, performance enhancements, client APIs and more. Show your support by starring or following the project!

Please read the following articles before submitting any contributions or filing an Issue:

 - [Contribution Guidlines](https://github.com/hewiefreeman/GopherGameServer/blob/master/CONTRIBUTING.md)
 - [Code of Conduct](https://github.com/hewiefreeman/GopherGameServer/blob/master/CODE_OF_CONDUCT.md)
