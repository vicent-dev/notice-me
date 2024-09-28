# notice-me

### ⚠️ Project under construction ⚠️ 

demo site: https://noticeme.vicentdev.com

Notice-me is a microservice that allows you to implement 
frontend async notifications from your backend system/s via websockets.

The value added by this project is mostly in the backend since the frontend implementation is not a lot more than a 
websocket connection but a frontend package to simplify the integration would be considered if needed.

Create notifications from frontend or other backend service and notify to all clients available.
You can also discriminate by client id or group id. (demo site group id feature not implemented in the frontend yet).

Currently, you can only create notifications via http but AMQP creation might be added in the future.
The notifications are persisted on the http request but the "notify" action is published into a RabbitMQ queue. 
When the notification is consumed from RabbitMq and published to clients through websocket it will be updated the field `notified_at`.

