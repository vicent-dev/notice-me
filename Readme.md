# Notice-me

### ⚠️ Project under construction ⚠️ 

demo site: https://noticeme.vicentdev.com

Notice-me is a microservice that allows you to implement 
frontend async notifications from your backend system/s via websockets.

The value added by this project is mostly in the backend since the frontend implementation is not a lot more than a 
websocket connection but a frontend package to simplify the integration would be considered if needed.

Create notifications from frontend or other backend service and notify to all clients available.
You can also discriminate by client id or group id.

This would be a very basic data flow with Notice-me:
- frontend client connects to Notice-me websocket.
- user does something in your system that will trigger eventually a notification. (e.g. complex import process of data that will probably take some time to finish).
- notification is created in Notice-me via http POST request.
- eventually your system will finish your process related with the notification  (e.g. all the data is finally imported in your system).
- publish from your system into queue `notification.notify` with the notification ID that you will recieve from create endpoint.
- Notice-me will consume that message and broadcast that notification to all fronted clients that matches the configured criteria.


## EntryPoints

There is two basic ways to interact with Notice-me:

- **http**: endpoints are for the websocket connection and for the CRUD of notifications (swagger documentation pending).
- **amqp**: one consumer for `notification.notify` queue which objective is to notify a notification and marked it as notified.

At the moment the demo just reads the pending notifications and publish them into the notification queue which means that 
the notifications will be notified when the backend demo is executed.
