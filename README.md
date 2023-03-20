# Go Server

## Description

A socket server built in Go which handles concurrent requests.  The application basically accepts push/pop requests for various payloads and pushes/retrieves them from a stack. It handles requests
in the order in which they come, however, it will also priortize the requests.
If a request comes into the application and the amount of processing it requires takes a long time, then any subsequent, faster requests will be processed first.  A supplmemental queue is used to handle the active requests in the system and to help prioritize which data should be processed accordingly.

## Instructions

The program is written in Go.  I used Go version 1.19, but I think older versions are compatible, too.

cd to the root directory, run:

```go run .```

Wait until server reports that it has started on port 8080

## Where to start reviewing the code
* Starting to read the code from the server.go file would be a good starting place.  This gives a high-level overview of how the application works.  From there, subsequent packages can be reviewed.

## What I learned
* I learned more about the Go language.  I chose it because it seemed to be good at handling threads of execution and has built in functionality for working with sockets.  I learned something new about Go almost every day I worked on the project.
*  I learned how to implement block/waiting using Channels in Go for effecient waiting (as opposed to busy looping).
* I learned that detecting a closed-socket connection can be tricky.  There were some instances where the socket server was successfully writing to a client that had already been closed.  I learned this is a soft close.  I believe these have been fixed by continuously polling every 10 ms to check for the closed socket connection.

## What I would like to improve
* I'd like to figure out a way to use 1 queue instead of 2.  The application uses 2 queues to handle both the push and pop requests.
* I'd like to improve the code which looks up the oldest socket connection in the connection manager (this is done to dismiss connections older than 10 seconds to prevent deadlock).  I think this could be more effecient, but I needed to timebox myself.
* There is no synchronization on the stack push/pop methods.  I put a lock instead on the code surrounding the calls to these operations because of some race conditions that were encountered.  I didn't want doubly nested mutex locks, so I removed them in the stack methods.  Eventually I'd like to have the entire stack be thread safe.