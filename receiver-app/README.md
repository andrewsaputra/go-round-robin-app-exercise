# Receiver API

Simple http api which receives http json request from client and echo the exact same payload as its response

## API Endpoints

| Path | Method | Payload |Description |
| --- | --- | --- | --- |
| `/status` | GET | - | Return HealthCheck status of the application |
| `/echojson` | POST | string | Receives request and echo back the payload as its response. Request payload must be a valid json string. |

## Usage 

### Running Application From Project

Use this command to run an instance of application in a custom port
```
go run . <target port>
```
e.g. : `go run . 4001`

Default port `4000` will be used if target port is not specified.

Optionally you can set environment variable `GIN_MODE=release` to reduce logging verbosity of the application's http framework, e.g. : `env GIN_MODE=release go run . 4001`


### Running Application From Binary

To build app binary use the following command
```
go build -o <app name>
```
e.g. : `go build -o myapp`

Then you can use the following command to run an instance of the application 
```
./<app name> <target port>
```
e.g.: `env GIN_MODE=release ./myapp 4000`

### Invoking API
Request Examples :

- `curl localhost:4000/status`
```
{"Status":"Healthy","StartedAt":"Tue, 17 Oct 2023 17:38:02 +0700"}
```

- `curl -X POST localhost:4000/echojson -d '{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}'`
```
{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}
```
