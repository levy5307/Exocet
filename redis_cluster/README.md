redis_cluster
===============

redis cluster  with sentinel


# how to start the cluster on one host

* edit script/cluster_load.sh and change the address(ip&port) of meta & master & slave & sentinel

* gen the cluster by cluster.sh: sh cluster.sh gen 2 100m

		the parameter 2 means cluster.sh should generate 2 redis instances and the parameter 100m means the instance's max memory size is 100MB.

* and then you gotta the cluster directory, step into this dir and start the cluster as follows

		sh cluster_load.sh start 2

* if you wanna shutdown the cluster, do it as follows

		sh cluster_load.sh stop 2