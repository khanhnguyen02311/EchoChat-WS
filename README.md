# EchoChat-WS
A module for EchoChat that allows for WebSocket connections.

## Message format
The input message format is as follows:
```json
{
  "type": "message",
  "data": {
    "group_id": "be59abbe-9976-11ee-b70d-a4423bfe9228",
    "type": "Message", // type must be one of ["Message", "File"]
    "content": "Message content or filename goes here"
  }
}
```

The response message format is as follows:
```json
{
  "type": "response",
  "status": "success", // ["success", "error"]
  "message": null,
  "notification": null,
  "content": "Error or success message goes here"
}
```

The new notification message format is as follows:
```json
{
  "type": "notification",
  "status": "new",
  "message": null,
  "notification": {
    "type": "GroupEvent",
    "time_created": "2023-12-13T05:45:33.894150312Z",
    "group_id": "be59abbe-9976-11ee-b70d-a4423bfe9228",
    "accountinfo_id_sender": 3,
    "content": "Notification content goes here"
  }
}
```
