# Spreadsheet back-end

## Test

The tests are executed during container build as a part of the `Dockerfile`.

To execute tests separately run:
```
docker build --target test --progress plain --no-cache .
```

Also you may execute tests locally by running:

```
go test -v ./...
```

## Run

To start application simply run:

```
docker compose up
```

## REST operations

```
# Set cell to a constant
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "123"}'

# Set cell to a formula expression
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "=var1*2"}'

# Get cell
curl localhost:8080/api/v1/devchallenge-xx/var1

# Get whole spreadsheet
curl localhost:8080/api/v1/devchallenge-xx
```

## Corner cases

### Cell identifies

You could use utf8 characters for `cellId`, the identifier should start with a
letter, but next characters could be any printable except for `+-*/()`:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/拿 -d '{"value": "2"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/á._ -d '{"value": "3"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/說 -d '{"value": "=á._+拿"}'
```

So `1abc` is not valid variable name as it starts with a number, not a letter.

Both cell and spreadsheet identifiers length is limited by the http protocol.

### Circular formulas

In case of circular dependency in formula result would be error. In a case of
such error the update formula would be rejected:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "1"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "=var2"}'

curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "=var1"}'
# Error occured because of circular dependency: var2 = var1 = var2 ...
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

Also division of `INTEGER` by `INTEGER` would result in `FLOAT` type for precision.

### Formula expression

You may specify unary operator right after a binary and it would be treated as
following token modifier, e.g. next formulas are legit: `=1 * -2` = -2,
`=1 + -2` = -1, `=1 + +2` = 3

### Division by zero

You will receive `ERROR` during formula evaluation if division by zero occurs:

```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "0.0"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "=1/var2"}'
```

```json
{
    "error": "division by zero",
    "result": "ERROR",
    "value": "=1/var2"
}
```

### Dependent cell breaking prevention

The system prevents from breaking any dependant cell to be broken. Consider next case:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/var3 -d '{"value": "0"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var2 -d '{"value": "=var3 - 1"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/var1 -d '{"value": "=1 / var2"}'

# Next UPSERT would be rejected as it would break var1
curl -X POST localhost:8080/api/v1/devchallenge-xx/var3 -d '{"value": "1"}'

# var3 is still 0
curl localhost:8080/api/v1/devchallenge-xx/var3
```

### Cell value type determination

Cell value that starts with `=` would be considered as the formula.

Sequence of digits optionally precedent with a plus or minus sign considered as
`INT` value: `1`, `-1`, `+1`

If a number has fractional part it would be treated as a `FLOAT`
value: `1.01`, `1.`, `-1.0`, `+1.0`

Every other value would be treated as the `STRING` value: `some string`, `1.0.0`,
`++1`

### Numbers size

Numbers size is not limited, next expression would evaluate correctly:
```
curl -X POST localhost:8080/api/v1/devchallenge-xx/byte -d '{"value": "255"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/short -d '{"value": "=byte * byte"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/int32 -d '{"value": "=short * short"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/int64 -d '{"value": "=int32 * int32"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/int128 -d '{"value": "=int64 * int64"}'
curl -X POST localhost:8080/api/v1/devchallenge-xx/whatever_you_want -d '{"value": "=int128 * int128 * int128 * int128 * int128"}'
```

```json
{
    "result": "3335910840989723103946716225008276488526215899027251552832106642768471211926062170159900171395553219118558904881722547998563003168023964769375828837692267436754178788760327734053134918212890625",
    "value": "=int128 * int128 * int128 * int128 * int128"
}
```
