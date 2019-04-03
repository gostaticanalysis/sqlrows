# sqlrows

[![CircleCI](https://circleci.com/gh/gostaticanalysis/sqlrows.svg?style=svg)](https://circleci.com/gh/gostaticanalysis/sqlrows)

`sqlrows` is a static code analyzer which helps uncover bugs by reporting a diagnostic for mistakes of `sql.Rows` usage.

## Install

You can get `sqlrows` by `go get` command.

```bash
$ go get -u github.com/gostaticanalysis/sqlrows
```

## QuickStart

`sqlrows` run with `go vet` as below when Go is 1.12 and higher.

```bash
$ go vet -vettool=$(which sqlrows) github.com/you/sample_api/...
```

When Go is lower than 1.12, just run `sqlrows` command with the package name (import path).

But it cannot accept some options such as `--tags`.

```bash
$ sqlrows github.com/you/sample_api/...
```

## Analyzer

`sqlrows` checks a common mistake when using `*sql.Rows`.

At first, you must call `rows.Close()` in a defer function. A connection will not be reused if you unexpectedly failed to scan records and forgot to close `*sql.Rows`.

```go
rows, err := db.QueryContext(ctx, "SELECT * FROM users")
if err != nil {
    return nil, err
}

for rows.Next() {
	err = rows.Scan(...)
	if err != nil {
		return nil, err // NG: this return will not release a connection.
	}
}
```

And, if you defer a function call to close the `*sql.Rows` before checking the error that determines whether the return is valid, it will mean you dually call `rows.Close()`.

```go
rows, err := db.QueryContext(ctx, "SELECT * FROM users")
defer rows.Close() // NG: using rows before checking for errors
if err != nil {
    return nil, err
}
```

It may cause panic and nil-pointer reference but it won't clearly teach you that is due to them.