package ark

// namespace deploys default Ark namespace (heptio-ark)
func namespace() string {
	return `
apiVersion: v1
kind: Namespace
metadata:
  name: heptio-ark
`
}

// serviceAccount creates ServiceAccount used by Ark pods
func serviceAccount() string {
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

// rbacRole is a ClusterAdmin RoleBinding allowing Ark to do backups and restores
func rbacRole() string {
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
