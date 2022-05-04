# Questions
The aim of the task is to see how you generally solve a problem, not tackle a specific programming problem. 
To be explicit, a running solution that does not fulfill all the requirements is better than perfect code that doesn’t fulfill any requirements.

# Questions that have to be answered (in the presentation)
- [ ] How did you interpret the task?
    - [ ] What questions came up?
    - [ ] What assumptions did you make?
- [ ] What are the limitations of your solution?
    - [ ] What improvements would you add in the future?
- [ ] Did you find the task, and its requirements, clear?

## How did you interpret the task?
- As a real task
- An opportunity to show which is my working way
  - Stages:
    - Clarify task definition and boundaries
    - Documentation and research
    - Plan the Smallest Iteration
    - Implement
    - Iterate
  - Goal:
    - A proposed solution to the task
- Final implementation is not production ready indeed
  - Needs: CI Integration
  - Timers update
  - K8s manifest are defined for minikube development
    - ImagePullPolicy:Never...

### What questions came up?
- Single CRD by cluster
  - this was maybe the main question to clarify (I think is pretty obvious, but...)
- I would ask more about the real Prometheus Server scenario
  - know more about the real needs
  - expected use cases
- Technically a few
  - I didn't know about finalizers
  - conciliation loop, current implementation has lots of limitations
  - resource defintion....

### What assumptions did you make?
- Single CRD by cluster
- POC Vanilla Prometheus Server
  - No federation
- My own limitations
  - Kubebuilder
  - Usual K8s development practices
    - testing
    - helpers
    - KISS and YAGNI

## What are the limitations of your solution?
- Prometheus-Server limitations
  - Volatile volume (PVC alternative example)
  - Vanilla definition
    - more elaborated ones as federation not allowed
    - Server stack can include AlertManager too (example)
- Development limitations
  - No Kubebuilder
    - controller internals focus
    - narrow operator implementation
      - easy path
      - does not real wait
  - Hardcoded conciliation
    - FSM is static
  - Hardcoded resources
    - Needs to move to charts (helm)
  - Real feedback loop
    - should be coupled to /-/ready prometheus?
    - right now just ensures resource creation/deletion

### What improvements would you add in the future?
- Kubebuilder (standard code)
- Real conciliation loop
- Admission Webhook (crd real limits)
  - Validation Prometheus config, reject if required
- Move receipts from hardcoded to helm charts
  - generic deployer indeed
- Conciliation Loop behaviour
  - status timeouts
  - real flow control
    - error recovery procedures
    - resilient behaviour
  
## Did you find the task, and its requirements, clear?
- Clear enough to move ahead and complete full simple first iteration
- Development limitations were not blocking
  - Undefined scenarios as the resource update (rollout new Prometheus Server version example)
    - Executed as workaround, without requiring big app changes
    - Show implementation limits
      - partial rollouts
        - version rollout: new deployment
        - config rollout: configmap rollout nad Prometheus Server reload

## Outcome
- I enjoyed a lot
- I learnt quite a few things
  - finalizers
  - resync
  - toolset limitations
  - ...
