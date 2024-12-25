# "Nested Cubes" example for Retro Debugger WebSocket API client

# Description:

This example demonstrates how to use Retro Debugger WebSocket API client to draw nested cubes effect which is entirely calculated on the client side.

![alt text](assets/Nested-Cubes.png)

# Usage:

First of all, make sure that the WebSockets server is enabled, navigate to: `Settings -> Emulation -> WebSockets debugger server`, and tick it on.

![alt text](assets/EnableWSServer.png)

To run example, execute:
- Using Makefile: `make run`
- Using Go: `go run .`

To build executable binary:
- Using Makefile: `make build`
- Using Go: `go build -o bin/cubes .`

# Author:
- Artur 'Mojzesh' Torun

# License:
- MIT
