# Introduction to UCM Starlight

![ucm arch](https://github.com/raycarroll/ucm/assets/3717348/2f0870d1-b90d-4a3d-b2d2-6fd8440d0959)

## Setup guide

### Prerequisites:

- kubernetes (https://kubernetes.io/docs/setup/)
- kind (https://kind.sigs.k8s.io/)
- Kubernetes cluster and it's admin credentials for further actions 

## Locally

For local setup - additional requirement is Docker (https://www.docker.com/) and RabbitMQ instance (https://www.rabbitmq.com/docs/download)

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

## Deploy on a cluster (kind)

Useful UI tool - https://k8slens.dev/


The required files are located in: 
```
ucm/deployment/.
``` 
### Useful oc commands

1. To check the status of your pods

```
kubectl get pods -n _namespace_
``` 
2. To delete a deployment
```
kubectl delete deployment deployment_name -n your_namespace
```
3. To verify if RabbitMQ service is healthy and running
```
kubectl get svc -n your_namespace
```
4. To get the logs of a given pod
```
kubectl logs _podname_ -n _namespace_
```
5. To execute command in a pod:
```
kubectl exec -it _podname_ -- /bin/ bash
```

### Step to apply the deployment configurations:

First, we need to configure the volume, volumeclaim and rabbitmq:

```
kubectl apply -f ./volume.yaml -n _namespace_
```

```
kubectl apply -f ./volumeclaim.yaml -n _namespace_
```
```
kubectl apply -f ./deployment_rabbitmq.yaml -n _namespace_
```
### Then, we can start deployment of: deployment.yaml and deployment_starlight.yaml

```
kubectl apply -f _filename_ -n _namespace_
```

You should be able to see the following output (example):

```
deployment.apps/ucm-producer-deployment created 
```

## Once all pods are up, you can run the following command to copy the data in the input folder:

```
kubectl cp /some/path/to/inputfile . _podname_:/starlight/data/input/ -n _namespace_ 
```

## The Output data can be viewed in the /starlight/data/output directory. You can either exec into the pod to view or copy down to local directory:

```
kubectl exec --stdin --tty _podname_ -n _namespace_ -- /bin/bash
cd /starlight/data/output
```
OR
```
kubectl cp _podname_:/starlight/data/output/ /some/local/dir/ -n _namespace_ 
```
