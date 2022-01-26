# KubeOne for Developers

## Goals
* "Developer Batteries included"
* Developers should get enabled to start working on the Kubernetes Cluster without having to deal with infrastructure setup problems
* Independent KubeOne AddOns which can be enabled/disabled depending on the needs

## Non Goals
* Having the plugins available in KKP.
* Setting up Development Environments for specific programming languages.
* Being to opiniated about tooling. Eg zsh or fish should not be included.
* Installing opiniated IDE AddOns. Only the minimal set should be delivered.

## Motivation and Background
For Developers it is still hard to get started with Kubernetes. For getting into production mode they have to care about some infrastructure tasks.

For example:
* Getting an IDE up and running
* Connect to the Kubernetes Cluster
* Installing the Kubernetes Dashboard
* Having a Container Registry in Place
* Having a proper Git setup
* Having GitOps tooling up and running
* Having kubectl and other useful tools (kubectx, bash completion, ) installed
* Having a tool in place for handling sensitive data

## Implementation

### IDE
* Using [Eclipse Che](https://github.com/eclipse/che) as Cloud IDE. Eclipse Che is a workspace server integrating Theia as a default. 
* Theia is a nodejs application which provides the IDE. Workspace management has to be provided by Eclipse Che. 
* Eclipse Che is holding all the workspace data and all user-specific file-based configurations, eg the .gitconfig

### Terminal integrations
* Theia provides a terminal which can be configured eg .bashrc
* We should ensure
    * The terminal is connected to the Kubernetes Cluster the IDE is running in. The kubeconfig should restrict permissions to developer matters.
* kubectl has to be installed.
* Either docker or crictl has to be installed.
* [Helm](https://github.com/helm/helm) has to be installed
* Useful tools like [kubectx](https://github.com/ahmetb/kubectx) have to be installed.
* [Stern](https://github.com/wercker/stern) should be installed
* Questionable candidates: Istioctl
* Git has to be installed. May we can support developers here also on credential management.

### Security
* The IDE has to be secured. Everything should be done via HTTPS (LetsEncrypt?) 
* Password Protection: Otherwise we create a huge security hole for the Kubernetes Cluster.
* Dex as IDP with Github

### Visualization of the Kubernetes Cluster
* Either via enabling the Kubernetes Dashboard or via the [VSCode Kubernetes AddOn](https://marketplace.visualstudio.com/items?itemName=ms-kubernetes-tools.vscode-kubernetes-tools) (if that is doable in Theia)

### Container Registry
* Open Source Container Registry should become its own KubeOne AddOn. Harbor should be supported, may other Open Source Container Registries can be considered.

### GitOps
* FluxCD and/or ArgoCD should get supported. They should become its own KubeOne AddOn. Their controllers should get installed into the KubeOne cluster. We should take care that the Git Repo holding the yaml files for the Kubernetes Cluster can be configured easily.

### Vault
* Vault should become its own KubeOne AddOn. Developers should have the possibility to store sensitive data. Attention has to be paid to the persistence of this data.

## Alternatives Considered

## Task & Effort

Rough Estimate: 
* [3] IDE
* [1] Terminal integrations
* [3] Security
* [1] Visualization of the Kubernetes Cluster
* [3] Container Registry
* [3] GitOps
* [2] Vault
-----------
16 MD