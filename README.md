# rabbitautoscaler

A autoscaler POC that will scale a deployment based on queue depth

## Container Image

Run the following to build and publish the rabbit auto-scaler container image:

```sh
podman build -t ghcr.io/tonygilkerson/rabbitas:v1 .

podman push ghcr.io/tonygilkerson/rabbitas:v1
```

## Rabbit Secret

The deployment in this chart needs a secret that looks something like the following for accessing the message broker. In my case I am deploying into a namespace that has a rabbitmq cluster created by the rabbitmq-operator.  The operator creates a default user secret. If you are not using the operator just create a secret manually that points at your rabbit instance.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: zoomq-default-user
  namespace: rabbitmq
type: Opaque
data:
  connection_string: REDACTED
  default_user.conf: REDACTED
  host: zoomq.rabbitmq.svc
  password: REDACTED=
  port: "5672"
  provider: rabbitmq
  type: rabbitmq
  username: default_user_Isn6VzzWwxRunYemAd1=
```

## Deploy App

```sh
helm upgrade rabbitas charts/rabbitas -n rabbitmq -i
```
