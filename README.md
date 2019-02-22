# Mattermost Push Notifications Recorder plugin

Query to get the PushNotifications with the enqueue/send/ack datetimes.

```sql
SELECT PN.*, EPN.enqueued AS enqueued FROM PushNotifications AS PN LEFT JOIN EnqueuedPushNotifications AS EPN ON PN.id = EPN.Id;
```
