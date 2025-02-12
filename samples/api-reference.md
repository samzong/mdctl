# API Reference v2.1

## Authentication

### JWT Authentication
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

**Token Acquisition:**
```http
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET
```

## Endpoints

### User Management

#### List Users
```http
GET /api/v1/users?page=2&per_page=20&fields=id,username,email&sort=-created_at
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | integer | 1 | Page number |
| per_page | integer | 10 | Items per page (max 100) |
| fields | string | - | Comma-separated field list |
| sort | string | created_at | Sort field (+asc, -desc) |

**Success Response:**
```json
{
  "data": [
    {
      "id": "usr_123",
      "username": "johndoe",
      "email": "john@example.com",
      "created_at": "2024-03-20T08:00:00Z",
      "updated_at": "2024-03-20T09:30:00Z"
    }
  ],
  "meta": {
    "current_page": 1,
    "total_pages": 5,
    "total_count": 42
  }
}
```

### Error Codes
| Status | Code | Scenario |
|--------|------|-----------|
| 401 | INVALID_TOKEN | Expired/malformed token |
| 403 | RATE_LIMITED | API rate limit exceeded |
| 422 | VALIDATION_ERROR | Request validation failed |
| 429 | TOO_MANY_REQUESTS | Rate limit exceeded |
| 503 | MAINTENANCE_MODE | System maintenance |

## WebSocket Enhancements

### Connection Management
```javascript
const RECONNECT_INTERVAL = 5000; // Reconnect delay

function connectWebSocket() {
  const ws = new WebSocket('wss://api.example.com/v1/ws');

  ws.onclose = () => {
    setTimeout(connectWebSocket, RECONNECT_INTERVAL);
  };

  ws.onerror = (err) => {
    console.error('WebSocket error:', err);
  };
}

// Heartbeat
setInterval(() => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'ping' }));
  }
}, 30000);
```

## Advanced Queries

### GraphQL Endpoint
```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    username
    posts(first: 5) {
      edges {
        node {
          title
          views
        }
      }
    }
  }
}
```