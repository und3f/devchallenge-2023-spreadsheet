# Spreadsheet back-end

## Test

The tests are executed during container build as a part of the `Dockerfile`.

To execute tests separately run:
```
docker build --target test .
```

Also you may execute tests locally by running:

```
go test ./...
```

## Run

To start application simply run

```
docker compose up
```

## REST operations


```
# Set cell
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "123"}'

# Set formula cell
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "=var1*2"}'

# Get cell
curl localhost:8080/api/v1/devchallenge-xx/var1

# Get spreadsheet
curl localhost:8080/api/v1/devchallenge-xx
```

## Limitations and edge cases

### Cell identifies

You could use utf8 characters for `cellId`:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/é -d '{"value": "123"}'
```

### Circular formulas

In case of circular dependency in formula result would be error:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "1"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "=var2"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "=var1"}'
```

Outputs: `{"value":"=var2","result":"ERROR"}`

Same applies for self referencing formulas:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "=var1"}'
```

### Types extrapolation

If `INTEGER` is mixed with the `FLOAT` the result would be `FLOAT`:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "=1 + 2.2"}'
```
Result: `{"value":"=1 + 2.2","result":"3.2"}`

`STRING` is not applicable to mathematical operations, still it could be placed in brackets: 
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "Some string"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "=(var1)"}'
```
Result: `{"value":"=(var1)","result":"Some string"}`
