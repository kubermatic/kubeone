output "kubeone_api" {
  value = {
    endpoint = "${digitalocean_loadbalancer.control_plane.ip}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name          = "${var.cluster_name}"
      cloud_provider        = "digitalocean"
      private_address       = "${digitalocean_droplet.control_plane.*.ipv4_address_private}"
      public_address        = "${digitalocean_droplet.control_plane.*.ipv4_address}"
      ssh_agent_socket      = "${var.ssh_agent_socket}"
      ssh_port              = "${var.ssh_port}"
      ssh_private_key_file  = "${var.ssh_private_key_file}"
      ssh_user              = "root"
    }
  }
}

output "kubeone_workers" {
  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    fra1-1 = {
      droplet_size        = "${var.droplet_size}"
      region              = "${var.region}"
      sshPublicKeys       = ["${digitalocean_ssh_key.deployer.public_key}"]
    }
  }
}
