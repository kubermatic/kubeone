output "kubeone_api" {
  value = {
    endpoint = "${aws_lb.control_plane.dns_name}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      public_address      = "${aws_instance.control_plane.*.public_ip}"
      private_address     = "${aws_instance.control_plane.*.private_ip}"
      user                = "ubuntu"
      ssh_public_key_file = "${var.ssh_public_key_file}"
      ssh_port            = "${var.ssh_port}"

      # ssh_agent_socket = "/run/user/1000/keyring/ssh"
    }
  }
}
