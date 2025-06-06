# Tender Backend

This service implements a simple HTTP API for managing tenders and bids. Data is stored in a local `data.json` file.

## Usage

Set `SERVER_ADDRESS` if you want to change the listening address (default `0.0.0.0:8080`).

Run the server:

```sh
go run ./...
```

Implemented endpoints:

- `GET /api/ping`
- `GET /api/tenders`
- `POST /api/tenders/new`
- `GET /api/tenders/my?username=USER`
- `GET|PUT /api/tenders/{id}/status`
- `PATCH /api/tenders/{id}/edit`
- `PUT /api/tenders/{id}/rollback/{version}`
- `POST /api/bids/new`
- `GET /api/bids/my?username=USER`
- `GET /api/bids/{tenderId}/list`
- `GET|PUT /api/bids/{id}/status`
- `PATCH /api/bids/{id}/edit`
- `PUT /api/bids/{id}/submit_decision?decision=...`
- `PUT /api/bids/{id}/feedback?bidFeedback=...`
- `PUT /api/bids/{id}/rollback/{version}`
- `GET /api/bids/{tenderId}/reviews?authorUsername=...`
