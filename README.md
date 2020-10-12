# Turing

Turing is a fast, scalable and extensible system that can be used to design, deploy and evaluate ML experiments in production. 

## Getting Started

The main components under turing can be visualised as follows. Refer to the README under the individual directories for getting started with the respective component.
```
.
├── api // Turing API
├── engines
│   ├── experiment // Turing Experiment Engine Interfaces
│   └── router // Turing Routing Engine
├── scripts
│   ├── fluentd-bigquery // Custom image for batch outcome logging to BQ
│   └── vertagen // Utility script for versioning
└── ui // Turing UI
```

## Local Deployment

To quickly deploy Turing locally into Docker, run:
```shell script
make start
```

The following command will build Turing from sources and deploy Turing with all necessary components 
into a local Docker cluster.  
Then, you can access Turing at [`http://localhost:8080/turing`](http://localhost:8080/turing)
   
## Architecture Overview

The following diagram shows the high level overview of the Turing system. Users
configure Turing routers from the Turing UI. Turing API creates the required
workloads and components to run Turing routers. The Turing router will be
accessible from the Router endpoint after a sucessful deployment.

![Turing architecture](./docs/assets/turing_architecture.png)


## Internal Migration Checklist

- [x] API Go module
- [x] Engines/router Go module
- [x] Engines/experiment Go module
- [ ] UI React app
- [x] Unit tests with GitHub Actions
- [ ] End-to-end test
- [ ] Load tests
- [ ] Docs
- [ ] Helm charts
