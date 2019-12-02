# ETCD Lease Watcher

A Basic example on how to use the golang etcd clientv3 to:

* Watch for deleted keys
* Find out if it was deleted because of a lease expiry
* Get the value of the key that was just deleted
* Starts a local etcd server from code
* Run all of this in a unit test
