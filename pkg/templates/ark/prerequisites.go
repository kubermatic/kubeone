package ark

// arkNamespace deploys default Ark namespace (heptio-ark)
func arkNamespace() string {
	return `
apiVersion: v1
kind: Namespace
metadata:
  name: heptio-ark
`
}

// arkServiceAccount creates ServiceAccount used by Ark pods
func arkServiceAccount() string {
	return `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ark
  namespace: heptio-ark
  labels:
    component: ark
`
}

// arkRBACRole is a ClusterAdmin RoleBinding allowing Ark to do backups and restores
func arkRBACRole() string {
	return `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: ark
  labels:
    component: ark
subjects:
  - kind: ServiceAccount
    namespace: heptio-ark
    name: ark
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
`
}
