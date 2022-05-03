# Prometheus Operator

## Task Description
We would like you to implement a Kubernetes Operator which creates a Prometheus server in a Kubernetes cluster when a Prometheus Custom Resource is created - essentially, a stripped down “demo” version of the existing prometheus-operator.

## Task Subject
The aim of the task is to see how you generally solve a problem, not tackle a specific programming problem.
To be explicit, a running solution that does not fulfill all the requirements is better than perfect code that doesn’t fulfill any requirements.

### Requirements
- [ ] When a Prometheus Custom Resource is submitted to a cluster, the operator should ensure that a matching Prometheus server is created, running inside the cluster.
- [ ] The Custom Resource should specify the Prometheus server version.
- [ ] The Prometheus server created should scrape metrics from the implemented Kubernetes Operator.
  - [ ] The Prometheus server’s scrape configuration should be configurable through the Custom Resource.
- [ ] When the Custom Resource is removed, the Prometheus server is removed too.

### Questions that have to be answered (in the presentation)
- [ ] How did you interpret the task? 
  - [ ] What questions came up?
  - [ ] What assumptions did you make?
- [ ] What are the limitations of your solution?
  - [ ] What improvements would you add in the future?
- [ ] Did you find the task, and its requirements, clear?


### Tips for the presentation
While we don’t want to force a certain structure to your presentation, we have had good experiences with the following running order:
- Short introduction about you as a person (~5m)
- Recap of the task (~5m)
- Talk about the questions above (~5m)
- Demo (~5m)
  - Protip: Start the demo when you start giving the presentation, or have a backup video at hand
  - Show us your design, architecture, and some code (~10m)
- Technical discussion / questions from the audience
  - You should have questions for the audience!
- Free discussion


```
vendor/k8s.io/code-generator/generate-groups.sh all github.com/marcosQuesada/prometheus-operator/pkg/crd/generated github.com/marcosQuesada/prometheus-operator/pkg/crd/apis "prometheusserver:v1alpha1" --go-header-file ./hack/boilerplate.go.txt --output-base "$(dirname "${BASH_SOURCE[0]}")/" -v 10 
```

```
docker build -t prometheus-operator . --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)

```

### TODOs
- timers
- project structure
- minikube restrictions (ImagePullPolicy...)
- metrics
- eventRecorder
- worker pool size
- OwnerRef
- admission webhook (limit to 1 CRD, reject all others)
- rise up unrecoverable errors
  - no resources in cluster example
  - those can be implemented on reconciliation loop as state jumps too
- Prometheus-Server limitations
  - Volatile volume
  - Vanilla definition
    - more elaborated ones as federation not allowed
- No Kubebuilder
  - controller internals focus 
  - narrow operator implementation
    - easy path
    - does not real wait
- Real feedback loop
  - should be coupled to /-/ready prometheus?
  - right now just ensures resource creation/deletion
    - not starting to serve

### Notes
Periodic resync
```
  if newPs.ResourceVersion == oldPs.ResourceVersion {
      // Periodic resync will send update events for all known prometheus servers.
      // Two different versions of the same prometheus servers will always have different RVs.
      return
  }

```
Fine Grained Updates
```
		
		//
		//// @TODO: Further iterations handle separated cases
		//if old.Spec.Config != ps.Spec.Config {
		//	// Update ConfigMap && Call Prometheus reload command
		//	return nil
		//}
		//
		//if old.Spec.Version != ps.Spec.Version {
		//	// Patch/Update Current Deployment
		//	return nil
		//}
```