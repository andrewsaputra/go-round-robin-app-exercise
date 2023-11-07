# Routing API

API endpoint which receives http json request from client and then forward it to selected Receiver API host for processing, and afterwards return the result back to the client.

Host selection is done using round robin logic with handling for some of the common negative cases such as connection failures or timeout.

Typically we can utilize circuit breaker frameworks to help with the failure handlings, however for the purpose of this exercise, those will be done manually.

## API Endpoints

| Path | Method |Description |
| --- | --- | --- |
| `/status` | GET | Return HealthCheck status of the application |
| `/registerhost` | POST | Register new host to load balancer targets. |
| `/deregisterhost` | POST | Deregister host from load balancer targets. |
| `/*` | ANY | Receive requests and forward it load balancer target(s) using specified routing algorithm. |

## Usage 

### Running Application From Project

Use this command to run an application instance in default port `3000`
```
go run .
```


Optionally you can set environment variable `GIN_MODE=release` to reduce logging verbosity of the application's http framework, e.g. : `env GIN_MODE=release go run .`


### Running Application From Binary

To build app binary use the following command
```
go build -o <app name>
```
e.g. : `go build -o myapp`

Then you can use the following command to run an instance of the application 
```
./<app name>
```
e.g.: `env GIN_MODE=release ./myapp`

### Invoking API
Request Examples :

- `curl localhost:3000/status`
```
{"startedAt":"Wed, 18 Oct 2023 15:09:16 +0700","status":"Healthy"}
```

- `curl localhost:3000/registerhost -d '{"hostAddress" : "http://localhost:4001"}'`
```
{"message":"Successful registration"}
```

- `curl localhost:3000/deregisterhost -d '{"hostAddress" : "http://localhost:4001"}'`
```
{"message":"Successful deregistration"}
```

- `curl -X POST localhost:3000/echojson -d '{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}'`
```
{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}
```
