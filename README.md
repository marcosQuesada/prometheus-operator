# Prometheus Operator

## Overview
Prometheus Operator deployer operator, It deploys a basic Prometheus Stack at Kubernetes Cluster based on a New CRD called Prometheus Server, which defines Prometheus Version and config.

Deployer operator will take care in all the required steps to put the Prometheus Stack deployed and in service.

## Basics
This Operator has been done has a learning exercise, so that, a few trade-offs has been done, as:
- No Controller framework (not yet indeed), it's not using Kubebuilder or Operator-SDK, everything is done from scratch.
  - Why: FOr the sake of learning how internals works :)

## Current features
- [X] When a Prometheus Custom Resource is submitted to a cluster, the operator should ensure that a matching Prometheus server is created, running inside the cluster.
- [X] The Custom Resource should specify the Prometheus server version.
- [X] The Prometheus server created should scrape metrics from the implemented Kubernetes Operator.
  - [X] The Prometheus server’s scrape configuration should be configurable through the Custom Resource.
- [X] When the Custom Resource is removed, the Prometheus server is removed too.
Apart from those:
- CRD auto register
- Prometheus Operator support for rolling out new Prometheus versions or configs (once they are running, example: upgrade prometheus version)


## Current limitations
A few things feels wrong or incomplete at the end of the first iteration:
- Conciliation loop hardcoded states
  - The concept of 'desired state' is hardcoded in the internals
- No 'Chart's concept, controlled resources are hardcoded too in what's called ResourceEnforcers


As you see many things can be improved, basically seems that moving to Helm charts is the perfect step, enforcers will get out from resource coupling, and they will execute the desired chart receipt.

### Prometheus Server Custom Resource Definition
Includes just 2 fields
- Prometheus version: docker official images at https://hub.docker.com/r/prom/prometheus/tags
- Prometheus config: raw config, no validation is done (one of the improvements points)
Status is handled as CRD Subresource
- CRD state progression events feds the conciliation loop

## Workflow
- The controller watches Prometheus Server CRDs (note that by definition, right now, 1 single CRD is expected in one unique cluster)
- Once a PrometheusServer has been created it will execute the conciliation loop as many times as required until having a full Prometheus Server stack deployed.
- Conciliation loop gets fed from K8s event updates, which limits the conciliation loop capabilities:
  - Unable to react on status timeouts
  - Fragile behaviour on unexpected situations
  - Segregating the control loop to an external one seems the clear path to win on resilient behaviour.

### Operator Scheme
As a regular operator, Prometheus Server controller watches its own CRD type, reacting to those events applying the conciliation loop.
![img.png](img.png)

### Resource Enforcer scheme (Prometheus Server stack)
Deployed Prometheus Stack is based on:
  - clusterRole & clusterRoleBinding
  - configMap
  - single instance Deployment
  - Service in top of Deployment

![img_1.png](img_1.png)

### Conciliation loop detail
Conciliation loop hardcode the 'Desired State' which is Running until a CRD termination signal is received, moment where CRD desired state is Terminated.
Implementation is polarized by FSM typical design, events are submitted to the FSM advancing its state and changing its behaviour.
Update Prometheus Server versions or configs are handled as full scheme redeploy, transitions to Reloading state, destroying all resources and jumping back to initialize all resources.
This behaviour can be improved a lot, examples:
- on Prometheus version change the procedure would be:
  - recreate Deployment resource pointing the new Prometheus Image version. Once completed rollout is done
- on Prometheus config change
  - recreate ConfigMap with the new configuration
  - reload Prometheus Server can be done without instance restart using /-/reload endpoint
  
![img_2.png](img_2.png)

## Further Iterations
Many improvements can be done:
- Kubebuilder: moving to standard code means dedicating time to our core business, in that case resilient controller
- Conciliation Loop Improvements: timeout by state detection, react to them
- Admission Webhook: this will limit and validate incoming CRDs (total CRD allowed number, validate Prometheus config before apply)
- Help charts usage, this probably will move the project to a generic system that it just deploys charts and conciliates them

## QuickStart
- Project developed using Cobra. Two entry points:
  - external: Allows external Kubernetes connectivity, useful for development purposes.
  - internal: Uses K8s internal client, the one to use to build real containers

### External run about
```
go run main.go external
```

## Development procedures
API CRD generation:
```
vendor/k8s.io/code-generator/generate-groups.sh all github.com/marcosQuesada/prometheus-operator/pkg/crd/generated github.com/marcosQuesada/prometheus-operator/pkg/crd/apis "prometheusserver:v1alpha1" --go-header-file ./hack/boilerplate.go.txt --output-base "$(dirname "${BASH_SOURCE[0]}")/" -v 10 
```

Build Docker as:
```
docker build -t prometheus-operator . --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)

```
Run test suite
```
go run --race main.go external
```

### Deploy procedures:
All receipts are developed with Minikube, a few limitations are there, as ImagePullPolicy: Never which is really useful for development.
On a dev/prod scenarios update manifest to a real behaviour.
- create monitoring namespace 
- apply rabc.yaml
- apply controller.yaml
Once done Prometheus Server CRD is registered in the cluster and our Operator will be watching it.


### Environment vars
- WORKERS: total workers consuming operator events queue
- RESYNC_INTERVAL: Shared informer resync period
- LOG_LEVEL: Logging level detail
- ENV: reflects deployment environment
- HTTP_PORT: Operator exposed http port
  - /metrics reports operator metrics
  - /healthz reports Release version Date and Hash Commit. Liveness probe endpoint
