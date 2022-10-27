disable_kubeapi_loadbalancer = true
subnets_cidr                 = 27
os                           = "rhel"

# Use smaller instances in Ireland for E2E tests
aws_region                = "eu-west-1"
control_plane_type        = "t3.medium"
control_plane_volume_size = 50
worker_type               = "t3.medium"
bastion_type              = "t3.micro"
