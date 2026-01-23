# API Response Envelope

Lightdash API endpoints consistently wrap their data in a standard JSON envelope.

## Success Structure

A successful response typically looks like this:

```json
{
  "status": "ok",
  "results": {
    "field1": "value1",
    "field2": "value2"
  }
}
```

- **`status`**: Usually `"ok"`.
- **`results`**: Contains the actual data for the requested resource.

## Implementation in Go

Each API operation should define a local response struct to handle this envelope:

```go
type GetExampleV1Response struct {
    Results models.Example `json:"results"`
    Status  string         `json:"status"`
}
```

The `Client` method should unmarshal into this struct and return only the `Results` field.

## Error Handling

The `doRequest` helper treats any non-2xx status code as an error. It includes the response body in the error message for debugging:

```go
return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
```

Common status codes to expect:

- `200 OK`: Successful GET/PUT.
- `201 Created`: Successful POST.
- `204 No Content`: Successful DELETE.
- `401 Unauthorized`: Invalid or missing API Key.
- `403 Forbidden`: Insufficient permissions.
- `404 Not Found`: Resource does not exist.
