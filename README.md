# Round Robin API

## Goal
Write a round robin API which receives HTTP Posts and routes them to one of a list of application APIs

## Requirements

#### Create Receiver API
- R1. API receives HTTP Post requests containing a JSON payload
- R2. API returns HTTP 200 response containing exact copy of the received request payload
- R3. Multiple instances of the API can be deployed, e.g. : using different ports in local machine
  - R3.1. At least 3 deployable instances are required for demo

#### Create Routing API
- R4. API receive HTTP Post requests from client
- R5. API routes the request to Receiver API instances and pass the response back to client
  - R5.1 Routing should be done using round robin approach
- R6. API should handle failure scenarios on the Receiver API targets, e.g. : server error, timeouts

## Usage Guide

TODO