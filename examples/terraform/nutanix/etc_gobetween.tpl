[api]
enabled = false

[servers.default]
protocol = "tcp"
bind = "0.0.0.0:6443"
balance = "roundrobin"
max_connections = 10000
client_idle_timeout = "10m"
backend_idle_timeout = "10m"
backend_connection_timeout = "2s"

[servers.default.discovery]
kind = "static"
static_list = [
    %{ for target in lb_targets ~}
    "${target}:6443",
    %{ endfor ~}
]

[servers.default.healthcheck]
kind = "ping"
interval = "10s"
timeout = "2s"
fails = 2
passes = 1
