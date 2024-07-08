# Introduction to UCM Starlight

![ucm arch](https://github.com/raycarroll/ucm/assets/3717348/2f0870d1-b90d-4a3d-b2d2-6fd8440d0959)

## Setup guide

### Prerequisites:

- oc https://docs.openshift.com/container-platform/4.8/cli_reference/openshift_cli/getting-started-cli.html
- OpenShift cluster and it's admin credentials for further actions 
- RabbitMQ

## Locally

For local setup - additional requirement is Docker (https://www.docker.com/)

1. Set up RabbitMQ and make sure that the credentials in producer.go and receive.go are used correctly according to local environment.

2. Build images of the processor, receiver and starlight from the given Dockerfiles. 

```
docker build -t username/image_name .
```

3. Run the images to get the functional result:

```
docker run username/image_name
```

From the producer should be seen the following result:

```
------------SENDING-------------
user  user
pass  pass
url  amqp://user:pass@host:5672/
2024/07/05 00:29:23 Num Files = 8
2024/07/05 00:29:23 File = spectrum_0461.txt
2024/07/05 00:29:23  [x] Sent lambda flu
2024/07/05 00:29:23  Moving file to /processed dir
2024/07/05 00:29:23  Moved file to /docker/starlight/shared_directory/config_files_starlight/spectrum//processed/spectrum_0461.txt
2024/07/05 00:29:23 File = spectrum_xpos_00_ypos_00_NGC6027_LR-V.txt
```

From the receiver:

```
2024/07/05 00:32:26  ------------------ reveive() --------------------- 
user  user
pass  pass
url  amqp://user:pass@host:5672/
2024/07/05 00:32:26  [*] Waiting for messages. To exit press CTRL+C
2024/07/05 00:32:26  -------------------------- Iterating messages  -------------------------------
2024/07/05 00:32:26 Read Flag =  false
2024/07/05 00:32:26 Read Flag =  false
2024/07/05 00:32:26  -------------------------- Received a message: spectrum_0461.txt -------------------------------
Writing message to data file
```

## Deploy on a cluster

The required files are located in: 
```
ucm/deployment/.
``` 
### Useful oc commands

1. To check the status of your pods

```
oc get pods -n _namespace_
``` 
2. To delete a deployment
```
oc delete deployment deployment_name -n your_namespace
```
3. To verify if RabbitMQ service is healthy and running
```
oc get svc -n your_namespace
```

### Step to apply the deployment configurations:

First, we need to configure the volume, volumeclaim and rabbitmq:

```
oc apply -f ./volume.yaml -n _namespace_
```

```
oc apply -f ./volumeclaim.yaml -n _namespace_
```
```
oc apply -f ./deployment_rabbitmq.yaml -n _namespace_
```
### Then, we can start deployment of: deployment.yaml and deployment_starlight.yaml

```
oc apply -f _filename_ -n _namespace_
```

You should be able to see the following output (example):

```
deployment.apps/ucm-producer-deployment created 
```
