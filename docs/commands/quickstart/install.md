# Quickstart Install

The quickstart install command allows to install the landscaper (and optionaly an OCI registry) in a specified kubernetes cluster, such as a minikube, kind or garden shoot cluster. This is the quickest way to get the landscaper up and running.

## Prerequisites
- K8s cluster
- [Helm 3](https://helm.sh/docs/intro/install/)

## Usage
```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry
```
To install a specific version of the landscaper chart, use the `landscaper-chart-version` argument.

For more details on the cli usage, consult [landscaper-cli_quickstart_install reference](../../reference/landscaper-cli_quickstart_install.md).

### Interact with OCI registry
If you decide to install the OCI registry, you have to use port-forwading, since it will not be exposed externaly. 
You can forward the port 5000 of the registry pod to your localhost with the following commnad:
```
kubectl port-forward oci-registry-<pod-id> 5000:5000 -n landscaper
```
Afterwards, you can use the tools of your choice to push artifacts against the localhost:5000 registry url. 
TODO: verify special /etc/hosts domain name for docker push

### Landscaper Values
The landscaper values are used during the internal helm install of the landscaper chart. Therefore, all values from the chart can be specified. 

> ❗ If you use the `--install-oci-registry` flag, set `landscaper.registryConfig.allowPlainHttpRegistries = true`

A minimum working example goes as follows:
```yaml
landscaper:

  registryConfig: # contains optional oci secrets
    allowPlainHttpRegistries: true
    secrets: {}
#     <name>: <docker config json>

  deployers:
  - container
  - helm
#  - mock

```


