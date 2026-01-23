---
name: research-lightdash-api
description: Research and investigate Lightdash API endpoints to understand request/response schemas and behavior.
---

# Research Lightdash API

## Description

This skill is used to thoroughly research and investigate a Lightdash API endpoint. It involves researching official documentation and performing live verification against a running Lightdash instance to confirm the actual JSON schema, including request parameters and response bodies.

## Input

The user should provide:

1. **API Endpoint**: (e.g., `/api/v1/projects/:projectUuid/spaces`)
2. **HTTP Method**: (GET, POST, PUT, DELETE, PATCH)
3. **Documentation URL**: (Optional) Link to the official Lightdash API docs.

## Workflow

### 1. Documentation Research

- **Search**: Use `web_search` or `WebFetch` to find the endpoint details in the [Lightdash API Docs](https://docs.lightdash.com/api-reference/v1/introduction).
- **Analysis**:
  - Identify required and optional request parameters (path, query, body).
  - Note the expected response structure.
  - Check if it uses the standard [API Response Envelope](../implement-verified-lightdash-api-client/references/response_envelope.md).

### 2. Live Schema Verification

If you have access to a Lightdash instance and an API key:

- **Environment Setup**:
  - Check for `LIGHTDASH_API_KEY` and `LIGHTDASH_URL` (defaults to `https://app.lightdash.cloud`) in the environment or `.env` file.
  - If missing, inform the user: "To perform live verification, please provide `LIGHTDASH_API_KEY` and `LIGHTDASH_URL`."
- **Execution**:
  - Use `curl` to fetch the live response. This is the preferred method.

    ```bash
    # Example for GET request
    curl -X GET "${LIGHTDASH_URL}/api/v1/your/endpoint" \
        -H "Authorization: ApiKey ${LIGHTDASH_API_KEY}" \
        -H "Content-Type: application/json"
    ```

  - For POST/PUT requests with a body:

    ```bash
    curl -X POST "${LIGHTDASH_URL}/api/v1/your/endpoint" \
        -H "Authorization: ApiKey ${LIGHTDASH_API_KEY}" \
        -H "Content-Type: application/json" \
        -d '{"key": "value"}'
    ```

  - Alternatively, use the [Schema Verification Script](assets/verify_schema.go) if a Go-based execution is preferred:

    ```bash
    LIGHTDASH_API_KEY=your_key LIGHTDASH_URL=https://app.lightdash.cloud go run .claude/skills/research-lightdash-api/assets/verify_schema.go /api/v1/your/endpoint
    ```

- **Discrepancy Analysis**:
  - Compare the live JSON response with the documentation.
  - Document any undocumented fields, differences in data types, or unexpected behaviors.
  - For POST/PUT/PATCH requests, verify if the request body schema matches the documentation.

### 3. Response Analysis

- **Envelope Check**: Verify if the response is wrapped in the standard `{ "status": "ok", "results": ... }` envelope.
- **Model Mapping**: Draft the Go struct definitions (models) based on the verified schema, ensuring proper types (e.g., UUID strings, timestamps, nullable fields).

## Reference

- [Lightdash API Documentation](https://docs.lightdash.com/api-reference/v1/introduction)
- [API Response Envelope](../implement-verified-lightdash-api-client/references/response_envelope.md)

## Assets

- [Schema Verification Script](assets/verify_schema.go)
