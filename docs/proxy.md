# How to use KubeOne in a proxy environment

Create a KubeOne configuration as usual and add the following section:

```yaml
proxy:
  http_proxy: 'http://proxy.example.com'
  https_proxy: 'https://proxy.example.com'
  no_proxy: '127.0.0.1/8,localhost,*.local,10.10.10.0/24,10.20.0.0/16,10.254.0.0/16,172.25.0.0/16'
```

This causes KubeOne to configure any created docker daemon to use the specified proxies according to [this](https://docs.docker.com/network/proxy/). Just specifying the proxy results in *all* traffic being tunnel through the proxy, which results in a broken environment. Therefore multiple IP ranges and host names must be included in the no_proxy field (Docker allows ranges contrary to other tools). The example contains ranges that are required:

- `127.0.0.1/8,localhost,*.local`: In general (Loopback and local network)
- `10.10.10.0/24,10.20.0.0/16,10.254.0.0/16,172.25.0.0/16`: For [Kubermatic](https://github.com/kubermatic/kubermatic) installations.
- Additional names and ranges specific to your environment
