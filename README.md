# go-toolkit

Shared building blocks for Go HTTP services on the standard library: one error vocabulary, one JSON envelope, structured logging with correlation IDs, `net/http` middleware, and the security primitives an auth flow needs.

It exists so that every service I build answers errors the same way, logs the same way and handles credentials the same way. It is the foundation of [Muster](https://github.com/diagnosis/muster), which runs it in production.

```
go get github.com/diagnosis/go-toolkit/v3
```

Requires Go 1.25 or newer. No web framework: everything works with `net/http` and `http.Handler`.

## Packages

| Package | What it gives you |
|---|---|
| `apperr` | A typed error, `StatusErr`, that carries a stable code, the HTTP status, a message safe to show a user, and an internal message for logs. Constructors for the common cases: `BadRequest`, `NotFound`, `Forbidden`, `Conflict`, `Validation`, `InvalidCredentials`, `TokenError`, `TooManyRequests` and more. `AsStatusErr` unwraps one from any error chain. |
| `responder` | `JSON` and `Error`, which write every response in one envelope: `{data, correlation_id, timestamp}` on success, `{error: {status, code, message, details, correlation_id, timestamp}}` on failure. `Error` turns a `StatusErr` into the right HTTP status and hides everything else behind a 500. |
| `logger` | A `slog` wrapper. `Init(env)` picks the format, and `Info`, `Warn` and `Error` take a context and attach its correlation ID to every line. |
| `middleware` | `CorrelationID`, `RequestLogger`, `CORS`, per-IP `RateLimit` with `Retry-After`, `RequireAuth` and `OptionalAuth` around an auth function you supply, and `SSEWriter` for server-sent events. |
| `secure` | Argon2id password hashing, a JWT signer for access and refresh tokens, opaque refresh tokens stored as hashes, one-time codes, HMAC-SHA256 signing and verification, and user-info fetchers for Google and GitHub OAuth. |
| `validator` | A small field validator (`Required`, `Email`, `Password`, `MinLength`, `MaxLength`, `Min`) whose `Errors()` returns a `StatusErr` with one message per field, ready for `responder.Error`. |
| `mailer` | A `Mailer` interface with SMTP, Resend and ZeptoMail implementations. |

## Example

A handler that validates input, returns a typed error and writes the envelope:

```go
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	correlationID, _ := logger.GetCorrelationID(r.Context())

	v := validator.New()
	v.Required("title", in.Title)
	v.MaxLength("title", in.Title, 255)
	if verr := v.Errors(); verr != nil {
		responder.Error(w, verr, correlationID) // 400 with per-field details
		return
	}

	item, err := s.items.Create(r.Context(), in)
	if err != nil {
		logger.Warn(r.Context(), "create failed", "err", err)
		responder.Error(w, err, correlationID) // StatusErr -> its status; anything else -> 500
		return
	}
	responder.JSON(w, http.StatusCreated, item, correlationID)
}
```

A service returns errors that already know their status:

```go
if o.HostID != hostID {
	return nil, apperr.Forbidden("only the host can update this outing", "forbidden: host required")
}
```

Middleware wraps a mux from the inside out:

```go
var h http.Handler = mux
h = middleware.RateLimit(rate.Limit(10), 20, 5*time.Minute)(h)
h = middleware.CorrelationID()(h)
h = middleware.CORS([]string{"https://example.com"})(h)

// per route
requireAuth := middleware.RequireAuth(authFromCookie) // func(*http.Request) (userID string, err error)
mux.Handle("GET /api/me", requireAuth(http.HandlerFunc(handleMe)))
```

## Design choices

- **Two messages per error.** `Message` goes to the client; `InternalMessage` and the wrapped error go to the log. A handler can pass any error to `responder.Error` without leaking internals.
- **A stable `code` beside the HTTP status.** Clients branch on `code`, so two different 409s can be told apart.
- **Correlation IDs end to end.** The middleware reads or creates one per request, the logger attaches it to every line, and the responder returns it in the body and the `X-Correlation-ID` header.
- **Auth is a function you pass in.** `RequireAuth` doesn't know about cookies or headers; the service decides how a request proves who it is.
- **Secrets are stored as hashes.** Refresh tokens and one-time codes are generated here and compared against a stored hash, never stored raw.

## Versioning

Semantic versioning with Go module major-version paths. The current major is `v3`, so the import path ends in `/v3`. Tags `v2.0.0` and later are available; the latest is `v3.1.0`.

## Development

```
make build   # go build ./...
make test    # go test ./...   (117 tests)
make lint    # golangci-lint run
```

## License

MIT. See [LICENSE](LICENSE).
