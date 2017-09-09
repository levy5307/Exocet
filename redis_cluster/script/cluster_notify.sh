#!/usr/bin/env bash
# ******************************************************
# DESC    : cluster notify script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-04-01 20:21
# FILE    : cluster_notify.sh
# ******************************************************

cur=`date "+%Y-%m-%d-%H-%M-%S"`
#cur=`date "+%F %T"`
role=$2
if [ $role == "leader" ] ; then # observer
	echo "failover{cur=$cur, name=$1, role=$2, old-master=$4-$5, new-master=$6-$7}" >> /tmp/redis_failover.log
fi
