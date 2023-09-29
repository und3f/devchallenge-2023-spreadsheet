# Spreadsheet

## Test

The tests are executed during container build as a part of the `Dockerfile`.

To execute tests separately run:
```
docker build --target test .
```

Also you may locally execute tests by running:

```
go test ./...
```

## Run

```
docker compose up
```

## REST operations

Next operations are using `httpie` cli:

```
# Set cell
http localhost:8080/api/v1/devchallenge-xx/var1 value='3'

# Get cell
http localhost:8080/api/v1/devchallenge-xx/var1

# Get spreadsheet
http localhost:8080/api/v1/devchallenge-xx
```
