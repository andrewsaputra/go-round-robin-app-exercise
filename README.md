by [andrewsaputra](https://github.com/andrewsaputra) - 18 Oct 2023

# Round Robin API

## Goal
Write a round robin API which receives HTTP Posts and routes them to one of a list of application APIs

## Requirements

#### Create Receiver API
- R1. API receives HTTP Post requests containing a JSON payload
- R2. API returns HTTP 200 response containing exact copy of the received request payload
- R3. Multiple instances of the API can be deployed, e.g. : using different ports in local machine
  - R3.1. Can deploy to at least 3 instances for demo

#### Create Routing API
- R4. API receive HTTP Post requests from client
- R5. API routes the request to Receiver API instances and pass the response back to client
  - R5.1 Routing should be done using round robin approach
- R6. API should handle common failure scenarios on the Receiver API targets, e.g. : server error, timeouts

## Tech Stack
- [Go 1.20.6](https://go.dev/doc/install)
- [Gin Web Framework](https://gin-gonic.com/)
- [Testify](https://github.com/stretchr/testify) (testing helper)

## Repository Structure

| Directory | Description | Pull Requests | Documentation |
| --- | --- | --- | --- |
| receiver-app | Contain project codes for application hosting Routing API functionalities (`R1 - R3`). | [PR #1](https://github.com/andrewsaputra/go-round-robin-app-exercise/pull/1) | [Document](https://github.com/andrewsaputra/go-round-robin-app-exercise/blob/main/receiver-app/README.md) |
| routing-app | Contain project codes for application hosting Routing API functionalities (`R4 - R6`). | [PR #2](https://github.com/andrewsaputra/go-round-robin-app-exercise/pull/2) | [Document](https://github.com/andrewsaputra/go-round-robin-app-exercise/blob/main/routing-app/README.md) |

For overview simplicity, codes for both applications are available in this same repository.

If designing for actual deployment, it's most likely a better idea to create standalone repository for each of Receiver App and Routing App. That's because their modification and release cycle looks different and it will also facilitate more natural integration with CI/CD pipeline

## Running Applications
Detailed guide for each application is available on the relevant application directory

1. Spawn multiple (at least 1) terminals in your machine
2. `cd ${REPO_ROOT}/receiver-app` on each terminals
3. Start `Receiver API` instances by running the following command on each terminal. You can change the ports as desired, but make sure each instance utilize different ports.
```
go run . 4001
go run . 4002
go run . 4003
...
```
4. Adjust the contents of `Routing API`'s [config](https://github.com/andrewsaputra/go-round-robin-app-exercise/blob/main/routing-app/configs/handlerconfig.json) file with correct host addresses from previous step, and other desirable values
5. Spawn another terminal which we'll use to host `Routing API` instance
6. `cd ${REPO_ROOT}/routing-app` on the new terminal
7. Start `Routing API` instance by running the following command. Currently it's coded to always use port `3000`.
```
go run .
```

Both applications are now running. You can now use http request tools, e.g. : curl, postman, browser, etc to invoke calls to Routing API.

Command examples : 
- `curl localhost:3000/status`
```
{"startedAt":"Wed, 18 Oct 2023 15:09:16 +0700","status":"Healthy"}```
```
- `curl -X POST localhost:3000/routejson -d '{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}'`
```
{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}
```
