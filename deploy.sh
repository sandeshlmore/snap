k -n snap-system apply -f ./config/crd/bases/db.db_snapshots.yaml

k -n snap-system apply -f ./config/rbac/role.yaml 

k -n snap-system apply -f ./config/rbac/service_account.yaml 

k -n snap-system apply -f ./config/rbac/role_binding.yaml

k -n snap-system apply -f ./config/deployment.yaml
