# pinqloq Go Sample

A plain `net/http` sample application that will demonstrate how to integrate the `pinqloq` Go SDK into a backend service.

## Requirements

- Go 1.22 or later

## Run

```bash
go run .
```

Open [http://127.0.0.1:3300](http://127.0.0.1:3300).

The test lab page lets you trigger HTTP scenarios returning 200, 400, 401, 404, or 500 and inspect the raw response. Set the `PORT` environment variable to use a different port.

## Test

```bash
go test ./...
```

## Status

This scaffold has HTTP scenario endpoints and a browser test lab only. `pinqloq` SDK integration — automatic `net/http` middleware logging, manual events, redaction — is a follow-up step.

## Project standards

Shared conventions are vendored from [pinq-doq](https://github.com/pinqponq/pinqdoq) under `.pinq-doq/` and copied into `.claude/rules/`. See [CLAUDE.md](CLAUDE.md).

## License

MIT
