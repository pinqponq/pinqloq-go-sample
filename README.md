# pinqloq Go Sample

A plain `net/http` sample application demonstrating how to integrate the `pinqloq` Go SDK into a backend service.

It includes a browser-based test lab for automatic HTTP logging, manual structured events, and redaction. The application generates synthetic data only.

Your Pinqloq secret key and collection names are entered at runtime in the **Connect** card on the page, sent once to the local server, and held in its process memory for that run only — never written to disk, an `.env` file, or source control. Restart the server and you enter them again.

## Requirements

- Go 1.22 or later
- A Pinqloq account
- A Pinqloq project and secret key
- One collection for automatic HTTP logs
- One collection for manual events

## Dashboard setup

1. Sign in to the [Pinqloq dashboard](https://pinqloq.pinqponq.io).
2. Create a project and copy its secret key.
3. Create two collections. Suggested names are `pinqloq_go_test_http` and `pinqloq_go_test_manual`.
4. View delivered logs in the [Pinqloq log panel](https://pinqloq-panel.pinqponq.io).

Never place the secret key in frontend code, a mobile application, source control, or any file served to users.

## Install and import pinqloq

The `pinqloq` Go SDK is published as a tagged module: [pkg.go.dev/github.com/pinqponq/pinqloq-go-sdk](https://pkg.go.dev/github.com/pinqponq/pinqloq-go-sdk).
A plain `go build`/`go test`/`go run` (including a fresh clone) already resolves it — no vendoring,
submodule, or special environment variables needed. Any other Go backend can install it the same way:

```bash
go get github.com/pinqponq/pinqloq-go-sdk@v1.0.0
```

## Run

```bash
go run .
```

Open [http://127.0.0.1:3300](http://127.0.0.1:3300). Set the `PORT` environment variable to use a different port.

The app starts unconnected — the pinqloq middleware and the manual/redaction routes stay disabled (`/api/config` reports `configured: false`) until you fill in the **Connect** card in the browser.

The browser UI provides:

0. A **Connect** card — enter your secret key and the two collection names to arm the SDK for this run.
1. HTTP scenarios returning 200, 400, 401, 404, or 500 — captured automatically by the pinqloq `net/http` middleware.
2. Manual events at Debug, Information, Warning, Error, and Fatal levels via `Logger().Enqueue`.
3. Redaction tests — one endpoint redacts only the `taxNumber` field (`password` is redacted unconditionally by the SDK's built-in floor), the other redacts everything on the endpoint.

## Test

```bash
go test ./...
```

Automated tests never send data to the live service: no session is configured while testing, which disables the middleware and the manual/redaction routes (asserted to return `503`) without touching the network. The tests also assert `/api/config` never echoes a secret key back.

## Project standards

Shared conventions are vendored from [pinq-doq](https://github.com/pinqponq/pinqdoq) under `.pinq-doq/` and copied into `.claude/rules/`. See [CLAUDE.md](CLAUDE.md).

## License

MIT
