output "kubeone_api" {
  value = {
    endpoint = "${aws_lb.control_plane.dns_name}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      public_address  = "${aws_instance.control_plane.*.public_ip}"
      private_address = "${aws_instance.control_plane.*.private_ip}"
      ssh_user        = "ubuntu"
      ssh_port        = "${var.ssh_port}"

      # specify either the private key or the SSH agent socket for
      # KubeOne to use in order to connect to this host
      # ssh_private_key_file = "${var.ssh_private_key_file}"

      # You can prefix your socket with "env:" to point to an
      # environment variable, like "env:SSH_AUTH_SOCK", instead
      # of hardcoding the socket path
      # ssh_agent_socket = "/run/user/1000/keyring/ssh"
    }
  }
}
