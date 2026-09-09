# pinqloq Go Sample

A plain `net/http` sample application demonstrating how to integrate the `pinqloq` Go SDK into a backend service.

It includes a browser-based test lab for automatic HTTP logging, manual structured events, and redaction. The application generates synthetic data only. Your Pinqloq secret key stays on the server and is never exposed to browser code.

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

## Configure

Copy `.env.example` to `.env` and fill in your secret key and collection names:

```bash
cp .env.example .env
```

```env
PINQLOQ_SECRET_KEY=your-project-secret-key
PINQLOQ_HTTP_COLLECTION=pinqloq_go_test_http
PINQLOQ_MANUAL_COLLECTION=pinqloq_go_test_manual
PORT=3300
```

The `.env` file is ignored by Git and must never be committed. Without it, the app still runs — the pinqloq middleware and manual/redaction routes are simply disabled (`/api/config` reports `configured: false`).

## Run

```bash
go run .
```

Open [http://127.0.0.1:3300](http://127.0.0.1:3300).

The browser UI provides:

1. HTTP scenarios returning 200, 400, 401, 404, or 500 — captured automatically by the pinqloq `net/http` middleware.
2. Manual events at Debug, Information, Warning, Error, and Fatal levels via `Logger().Enqueue`.
3. Redaction tests — one endpoint redacts only the `taxNumber` field (`password` is redacted unconditionally by the SDK's built-in floor), the other redacts everything on the endpoint.

## Test

```bash
go test ./...
```

Automated tests never send data to the live service: `PINQLOQ_SECRET_KEY` is left unset while testing, which disables the middleware and the manual/redaction routes (asserted to return `503`) without touching the network.

## Project standards

Shared conventions are vendored from [pinq-doq](https://github.com/pinqponq/pinqdoq) under `.pinq-doq/` and copied into `.claude/rules/`. See [CLAUDE.md](CLAUDE.md).

## License

MIT
