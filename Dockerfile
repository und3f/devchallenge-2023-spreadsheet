FROM golang as base

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Execute tests
FROM base as test

RUN ["go", "test", "-v", "./..."]

# Build executable
FROM test as build

RUN go build -v -o /usr/local/bin/spreadsheet-backend ./cmd/service/main.go

# Run application
FROM golang as production

ENV REDIS_ADDR="localhost:6379"
EXPOSE 8080

COPY --from=build /usr/local/bin/spreadsheet-backend /usr/local/bin/spreadsheet-backend

CMD ["spreadsheet-backend"]
