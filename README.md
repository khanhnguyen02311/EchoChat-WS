# EchoChat-WS
A module for EchoChat that allows for WebSocket connections.

## Connection:
The WebSocket connection have this format: `wss://application-url.com:port/ws?token=abcxyz`

The connection requires a token to be sent in the query string.
- `token`: The EchoChat user's access token

```

## Input message format
The new message format is as follows:
```json
{
  "type": "message-new",
  "data": {
    "group_id": "00000000-0000-0000-0000-000000000000",
    "type": "Message", // type must be one of ["Message", "File"]
    "content": "Message content or filename goes here"
  }
}
```

The notification mark as read message format is as follows:
```json
{
  "type": "notification-read",
  "data": {
    "group_id": "00000000-0000-0000-0000-000000000000",
    "type": "GroupEvent", // type must be one of ["GroupEvent", "GroupRequest"]
  }
}
```

## Output message format
The response message format is as follows:
```json
{
  "type": "response",
  "status": "success", // can be one of ["success", "error"]
  "message": null,
  "notification": null,
  "content": "Error or success message goes here"
}
```

The new notification format is as follows:
```json
{
  "type": "notification",
  "status": "new",
  "message": null,
  "notification": {
    "type": "GroupEvent",
    "time_created": "2023-01-01T12:12:12.121212121Z",
    "group_id": "00000000-0000-0000-0000-000000000000",
    "accountinfo_id_sender": 1,
    "content": "Notification content goes here"
  }
}
```
