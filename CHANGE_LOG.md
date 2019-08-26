# Change Log
This is where all important changes and bug fixes will be described in detail. Each entry is labeled with a patch version number, and that is the version in which the changes and/or fixes were made. All prior versions will lack the described changes and/or fixes described.

> If you are experiencing any bugs or errors because of an update, please report them by [opening an issue](https://github.com/hewiefreeman/GopherGameServer/issues/new/choose)!

### Key
  :wrench: : Bug fix
  
  :warning: : Code-breaking change
  
  :newspaper: : New feature
  
  :monorail: : Optimization
<br><br>
## v1.0-BETA.2
  - :newspaper: Added a `version` macro to display current running server version
  - :monorail: :warning: ([commit](https://github.com/hewiefreeman/GopherGameServer/commit/941c558bfe44f237f150918187785cceb8aafecd)) Restoring logic has been simplified, but any previous version restore files will fail to restore!

## v1.0-ALPHA.5
  - :monorail: :warning: ([commit](https://github.com/hewiefreeman/GopherGameServer/commit/a5edd57bcc61fc6f5d10b194f4a433e7a5ed51da)) Merged the `rooms` and `users` packages into one `core` package. **Requires refactoring your server code)**

> Version **1.0-ALPHA.5** merged the packages `users` and `rooms` into a single package `core`. If you've updated your server from **1.0-ALPHA.4** or below, you will need to edit your code and replace any instance of `rooms` or `users` with `core`. This also changes how [Rooms](https://github.com/hewiefreeman/GopherGameServer/wiki/Rooms#blue_book-creating--deleting-rooms) are **created** and **retrieved** ([core.GetRoom](https://godoc.org/github.com/hewiefreeman/GopherGameServer/core#GetRoom)), and how [Users](https://godoc.org/github.com/hewiefreeman/GopherGameServer/core#GetUser) are **retrieved**.
