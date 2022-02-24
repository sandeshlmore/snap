# Deploy CRD
k apply -f deploy/crd.yaml


# Deploy Controller

k apply -f ./deploy/rbac/role.yaml 

k apply -f ./deploy/rbac/service_account.yaml 

k apply -f ./deploy/rbac/role_binding.yaml

k apply -f ./codeploynfig/deployment.yaml


# create CRD resource which will trigger snapshot of volume
k apply -f ./deploy/snapshot1.yaml



