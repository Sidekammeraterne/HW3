# HW3
---How to run the program---

Clients must be started manually in separate terminals.
All events are logged using go's log package.

--How To--
1. Start server
    Open a terminal, in root of project
    write `go run ./server`
    The server will start listening on port :5050    

2. Start a client 
    Open a terminal, in root of project
    write `go run ./client`
    repeat for each client in separate terminals

3. Write a message or leave the server
   A client can either leave system (write `leave`) or publish a message (write `message [insert message here]`)
             e.g. message hello
4. Stop server
   In the server terminal, Write `stop`

---Logging---
Server Logs include: Startup and shutdown events, Client join/leave events, Broadcast messages.
Client Logs include: Received broadcast messages and message delivery confirmations.
Logs are written to both Terminal (stdout) and logged in file `logFile.log`, cleared on every server start.
Log entry consists of: Timestamp, Component name (server/client), Event type & relevant identifiers. They end with the logical time.
e.g. 2025/10/27 12:05:33 [Server][Broadcast] to 2: Client 1 published "Hello" at logical time 4

-- Log ordering --
Because each client and the server run as independent processes, they all write concurrently to the same shared log file.
Since it is not guaranteed that writes from different processes will appear in exact order, each process buffers its output independently.
Therefore, even though all log entries include real timestamps, the physical order of lines in the file may appear  slightly out of sequence.
To construct a chronological view, the combined log file must be sorted lexicographically by timestamp after execution.
This can be done with the command `sort logFile.log > sorted.log`
