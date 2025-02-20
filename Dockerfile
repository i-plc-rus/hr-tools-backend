FROM golang:1.23.1-bullseye as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest \
    && swag init \
    && go build

FROM golang:1.23.1-bullseye as runner

WORKDIR /app

COPY --from=build /app/hr-tools-backend .
COPY --from=build /app/docs docs
COPY --from=build /app/static_preload static_preload
COPY --from=build /app/static static

CMD ["/app/hr-tools-backend"]