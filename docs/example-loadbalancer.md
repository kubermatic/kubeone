# Terraform loadbalancers in examples
For providers that doesn't support LoadBalancer as a Service functionality we
included a working example of how it might look like in your setup. Provided
example is not a requirement and you can always use your own solution.

## What software is used?
For those examples we use [gobetween][1] project. It is free, open-source,
modern & minimalistic L4 load balancer solutions that's easy to integrate into
terraform.

## But it's a SPoF (Single Point of Failure)!
Yes, it is. We provide this only as an example how it might look like, and at
the same time trying to stay minimal on resources. As provider you're using
doesn't support LBaaS, it's completely up to you how you organize your frontend
loadbalancing and HA for your kube-apiservers.

Possibilities to achieve truly HA loadbalancing is to bootstrap 2 of those LBs
and use one of the following:
* DNS to point to both machines.
* keepalived and VirtualIP if provider allows it.
* use some external software with predefined IPs and exclude gobetween bits from
  terraform entirely.

## What about my Haproxy/nginx/<your favorite proxy solution>?
As our example in terraform is exactly this — just an example, you are free to
use whatever else solution. Gobetween is not a requirement. The only requirement
would be to provide `apiEndpoint.host` (and optional `apiEndpoint.port`) in
configuration, or terraform outputs `kubeone_api.values.endpoint`.

## Can this be used as a loadbalancer for Ingress?
No, provided example loadbalancer solution only cares about kubernetes API
availability, it's not universal solution for all your workloads.

[1]: http://gobetween.io
