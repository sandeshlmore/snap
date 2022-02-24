/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbv1alpha1 "github.com/db/snap/api/v1alpha1"
	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SnapshotReconciler reconciles a Snapshot object
type SnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=db.db,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.db,resources=snapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.db,resources=snapshots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Snapshot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	var snap dbv1alpha1.Snapshot
	err := r.Get(ctx, req.NamespacedName, &snap)
	if err != nil {
		fmt.Printf("Error: unable to fetch snapshot for req: %s\n", req.NamespacedName)
	}

	pvcList := &corev1.PersistentVolumeClaimList{}
	filters := []client.ListOption{
		client.MatchingLabels{
			"snapshotpvcselector": snap.Spec.SnapShotPVCSelector,
		},
		client.InNamespace(snap.Namespace),
	}

	err = r.Client.List(ctx, pvcList, filters...)

	if err != nil {
		fmt.Printf("Error: Could not get Pvc list....\n")
	}

	for _, pvc := range pvcList.Items {
		fmt.Printf("Taking Snapshot for PVC %s\n", pvc.Name)
		err = CreateSnapshot(ctx, pvc, r)
		if err != nil {
			fmt.Printf("Error: Failed to create snapshot for pvc %s\n", pvc.Name)
			fmt.Println("Error: ", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1alpha1.Snapshot{}).
		Complete(r)
}

// Create New snapshot for a PVC
func CreateSnapshot(ctx context.Context, pvc corev1.PersistentVolumeClaim, r *SnapshotReconciler) error {
	// currentTime := time.Now().Format("2006-01-02-15-04-05.0000")
	currentTime := time.Now().Format("20060102150405")
	className := "csi-hostpath-snapclass" // TODO: make it dynamic

	snapTemplate := &snapshot.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", pvc.Name, currentTime),
			Namespace: pvc.Namespace,
		},
		Spec: snapshot.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &className,
			Source: snapshot.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvc.Name,
			},
		},
	}

	// fmt.Println("Snap template: ", *snapTemplate)
	err := r.Client.Create(context.Background(), snapTemplate)

	if err == nil {
		fmt.Printf("Volume Snapshot taken: %s-%s\n", pvc.Name, currentTime)
	}

	return err
}
