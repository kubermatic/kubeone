data "openstack_networking_network_v2" "external_network" {
  name     = "${var.external_network_name}"
  external = true
}
