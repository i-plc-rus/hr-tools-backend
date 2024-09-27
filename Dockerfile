FROM golang:1.23.1-bullseye as BUILDER

WORKDIR /app/
COPY go.mod ./
COPY go.sum ./

WORKDIR /app/

RUN go mod download
COPY . .

RUN go build -o ht-tools-backend main.go

FROM alpine:3.14 as RUNNER

WORKDIR /app/

COPY --from=BUILDER /app/ht-tools-backend ht-tools-backend
COPY --from=BUILDER /app/.env .env

RUN chmod +x ./ht-tools-backend

CMD ["/app/ht-tools-backend"]