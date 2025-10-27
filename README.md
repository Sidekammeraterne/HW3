# HW3
Clients must be started manually in separate terminals.
All events are logged using go's log package.

--Logs--
Server Logs include: Startup and shutdown events, Client join/leave events, Broadcast messages.
Client Logs include: Received broadcast messages and message delivery confirmations. 
Logs are written to both Terminal (stdout) and logged in file `logFile.log`, cleared on every server start.
Log entry consists of: Timestamp, Component name, Event type & relevant identifiers.
e.g. 2025/10/27 12:05:33 [Server][Broadcast] to 2: Client 1 published "Hello" at logical time 4

--How To--
1. Start server
    Open a terminal, in root of project
    write `go run ./server`
    The server will start listening on port :5050    

2. Start a client 
    Open a terminal, in root of project
    write `go run ./client`
    repeat for each client

   a client can either leave system (write `leave`) or publish a message (write `message [insert message here]`)
             e.g. message hello
3. Stop server
   Write `stop`
