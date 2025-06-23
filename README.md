# Broker Platform Backend

A Go + MongoDB backend for a multi-stock broker platform.

## 🛠 Tech Stack

- **Language:** Go  
- **Web Framework:** [Gin](https://github.com/gin-gonic/gin)  
- **Database:** MongoDB (via the official [mongo-driver](https://github.com/mongodb/mongo-go-driver))  
- **Auth:** JWT tokens (`golang-jwt/jwt/v4`)  
- **Circuit Breaker:** Sony [gobreaker](https://github.com/sony/gobreaker)  
- **Protobuf & gRPC:** `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`  

## Features

- **HTTP API** (Gin) on port `8080`  
- **gRPC API** on port `50051`  
- **grpc-gateway HTTP proxy** on port `8081`  
- **User Signup & Login** with JWT access + refresh tokens  
- **Protected endpoints** for Holdings, Orderbook, Positions  
- **MongoDB** persistence for users & refresh tokens  
- **Circuit breaker** on Mongo calls (Sony gobreaker)  
- **Protocol Buffers** definitions + **grpc-gateway** integration  

## 🚀 Quick Start

### 1. Clone & Install

```bash
git clone https://github.com/hahahamid/broker-backend.git
cd broker-backend
go mod download
```

### 2. Environment

Create a .env in project root:

```
MONGO_URI=mongodb://localhost:27017
DB_NAME=brokerdb
JWT_SECRET=supersecretkey
REFRESH_SECRET=anotherrefreshsecret
ACCESS_TOKEN_EXPIRE_MINUTES=10
```

### 3. Install Protobuf Compiler

### 4. Fetch Google APIs Protos

```bash
git clone https://github.com/googleapis/googleapis.git
```

### 5. Generate gRPC & Gateway Code

```
protoc \
  -I proto \
  -I googleapis \
  --go_out=proto --go_opt=paths=source_relative \
  --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=proto --grpc-gateway_opt=paths=source_relative \
  proto/broker.proto
  ```
## ▶️ Running

Make sure MongoDB is running locally on the URI in your .env.

```bash
go build ./cmd/server
./server 
```

- HTTP (Gin): http://localhost:8080

- gRPC: localhost:50051

- grpc-gateway: http://localhost:8081

## 🔍 API Endpoints

### Public Endpoints

| Method | Path      | Description          |
|--------|-----------|----------------------|
| POST   | `/signup` | Create new user      |
| POST   | `/login`  | Obtain JWT tokens    |
| POST   | `/refresh`| Refresh tokens       |
| GET    | `/health` | Health check         |

### Protected Endpoints (Require JWT)

| Method | Path          | Description                          |
|--------|---------------|--------------------------------------|
| GET    | `/holdings`   | Mock user holdings                   |
| GET    | `/orderbook`  | Mock past orders + PNL card          |
| GET    | `/positions`  | Mock active positions + PNL card     |

**Note:** Protected endpoints require the following header:
```http
Authorization: Bearer <ACCESS_TOKEN>
```

### 🧪 Testing

- ### HTTP (cURL/Postman)
```bash
# Signup
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret123"}'
# → { "access_token": "...", "refresh_token": "..." }

# Protected
curl http://localhost:8080/holdings \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

- ### grpc-gateway

Same HTTP calls → localhost:8081 instead of :8080.

- ### gRPC (grpcurl)

```bash
# Signup
grpcurl -plaintext localhost:50051 broker.Broker/Signup \
  -d '{"email":"you@example.com","password":"secret123"}'

# GetHoldings
grpcurl -plaintext \
  -H "authorization: Bearer <ACCESS_TOKEN>" \
  localhost:50051 broker.Broker/GetHoldings
  ```