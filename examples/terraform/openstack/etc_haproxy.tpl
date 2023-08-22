global
    chroot      /var/lib/haproxy
    pidfile     /var/run/haproxy.pid
    maxconn     4000
    user        haproxy
    group       haproxy
    daemon
    stats socket /var/lib/haproxy/stats

defaults
    log                     global
    retries                 3
    timeout queue           30s
    timeout connect         10s
    timeout client          1m
    timeout server          1m
    timeout check           10s
    maxconn                 3000

frontend k8s-control-plane
    bind *:6443
    default_backend k8s-control-plane

backend k8s-control-plane
    balance roundrobin
    mode tcp
    default-server maxconn 20
%{ for target in lb_targets ~}
    server ${target} ${target}:6443 check
%{ endfor ~}
