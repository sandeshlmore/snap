# Volume Snapshot Assignment


### Goal:

To write a Kubernetes controller/operator that is going to backup the volumes, using Kubernetes snapshot APIs, that are provided by the users as input. Once the volume is snapshotted users can also choose to restore that snapshot to get their data back.


### MVP:

If I install MySQl application, insert some data into the database, take backup of the volume and then delete the data from MySQL database. The data that we deleted from the database should be revoverable if we restore the volume snapshot that was taken.


### How it works:

***Backup***


- User can create a CRD for taking snapshots for  PVCs.

    CRD properties:
        
        snapshotpvcselector : PVC should have this label

- When CRD is created Controller takes Snapshot of the PVCs for which the label snapshotpvcselector is defined 


***Restore***

- Restore of Snapshot can be done by adding "datasource" field in PersistentVolumeClaim object.
e.g.
```
    dataSource:
        name: snap-shot-name
        kind: VolumeSnapshot
        apiGroup: snapshot.storage.k8s.io

```    

