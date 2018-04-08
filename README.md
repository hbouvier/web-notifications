# Web Push notification

## Build and Run

```bash
$ make
```

## Push a notification

First go to the web page bellow and register with the email `anonymous@acme.com`.

[Web registration form](http://localhost:8000/)

Then run this curl command to send a notificaiton

```bash
$ curl -v http://localhost:8000/api/v1/push -HContentType:application/json -d '{"subscriber":"anonymous@acme.com","event":{"title":"Yo","body":"Hello world2!","icon":"images/icon.png","badge":"images/badge.png","data":{"href":"https://www.google.com"}}}'
```
