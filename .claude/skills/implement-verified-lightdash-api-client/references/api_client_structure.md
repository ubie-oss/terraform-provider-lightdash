# API Client Structure

The Lightdash API client is implemented in `internal/lightdash/api`.

## `Client` Struct

The `Client` struct (defined in `client.go`) is the core of the API interaction. It manages:

- **`HostUrl`**: The base URL of the Lightdash instance.
- **`Token`**: The Personal Access Token for authentication.
- **`HTTPClient`**: A pre-configured `http.Client`.

## Authentication

Authentication is handled in `doRequest` by adding the `Authorization` header:

```go
req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", c.Token))
```

## Concurrency Limiting

To prevent overwhelming the Lightdash API, the client uses a `ConcurrencyLimitingTransport`. This is a middleware for `http.RoundTripper` that uses a buffered channel as a semaphore to limit in-flight requests (default: 100).

## Request Execution (`doRequest`)

All endpoint methods should use `c.doRequest(req)`. This helper:

- Adds mandatory headers (`Accept`, `Content-Type`, `Authorization`).
- Executes the request using the concurrency-limited transport.
- Reads and returns the response body as `[]byte`.
- Returns an error for non-successful HTTP status codes.
