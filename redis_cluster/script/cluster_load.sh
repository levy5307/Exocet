#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster cluster devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-04-06 16:41
# FILE    : cluster.sh
# ******************************************************

meta_master_ip="192.168.1.100"
meta_master_port="6000"

meta_slave_ip="192.168.1.100"
meta_slave_port="6001"

master0_ip="192.168.1.100"
master0_port="6379"

slave0_ip="192.168.1.100"
slave0_port="6380"

master1_ip="192.168.1.100"
master1_port="16379"

slave1_ip="192.168.1.100"
slave1_port="16380"

sentinel0_ip="192.168.1.100"
sentinel0_port="26380"

sentinel1_ip="192.168.1.100"
sentinel1_port="26381"

sentinel2_ip="192.168.1.100"
sentinel2_port="26382"

start_meta() {
	[[ -d meta_master ]] && cd meta_master && sh meta-master-load.sh start $meta_master_ip $meta_master_port && cd ..
	[[ -d meta_slave ]] && cd meta_slave && sh meta-slave-load.sh start $meta_slave_ip $meta_slave_port  $meta_master_ip $meta_master_port && cd ..
}

start_master() {
	idx=0 && [[ -d master$idx ]] && cd master$idx && sh master-load.sh start $master0_ip $master0_port && cd ..
	idx=1 && [[ -d master$idx ]] && cd master$idx && sh master-load.sh start $master1_ip $master1_port && cd ..
}

start_slave() {
	idx=0 && [[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh start $slave0_ip $slave0_port $master0_ip $master0_port && cd ..
	idx=1 && [[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh start $slave1_ip $slave1_port $master1_ip $master1_port && cd ..
}

start_sentinel() {
	notify_filename="/tmp/cluster_notify.sh"
	instance_set="meta:$meta_master_ip:$meta_master_port@cache0:$master0_ip:$master0_port@cache1:$master1_ip:$master1_port"
	idx=0 && [[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh start $sentinel0_ip $sentinel0_port $notify_filename $instance_set && cd ..
	idx=1 && [[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh start $sentinel1_ip $sentinel1_port $notify_filename $instance_set && cd ..
	idx=2 && [[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh start $sentinel2_ip $sentinel2_port $notify_filename $instance_set && cd ..
}

stop_meta() {
	[[ -d meta_master ]] && cd meta_master && sh meta-master-load.sh stop && cd ..
	[[ -d meta_slave ]] && cd meta_slave && sh meta-slave-load.sh stop && cd ..
}

stop_master() {
	idx=0 && [[ -d master$idx ]] && cd master$idx && sh master-load.sh stop && cd ..
	idx=1 && [[ -d master$idx ]] && cd master$idx && sh master-load.sh stop && cd ..
}

stop_slave() {
	idx=0 && [[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh stop && cd ..
	idx=1 && [[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh stop && cd ..
}

stop_sentinel() {
	idx=0 && [[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh stop && cd ..
	idx=1 && [[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh stop && cd ..
	idx=2 && [[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh stop && cd ..
}

usage() {
	echo "Usage: $0 {start|stop} {meta|master|slave|sentinel}"
	exit
}

if [ $# != 2 ]; then
	usage
fi
opt=$1
role=$2
case C"$opt" in
	Cstart)
		if [ "$role" = "meta" ]; then
			start_meta
		elif [ "$role" = "master" ]; then
			start_master
		elif [ "$role" = "slave" ]; then
			start_slave
		elif [ "$role" = "sentinel" ]; then
			start_sentinel
		else
			usage
		fi
		;;
	Cstop)
		if [ "$role" = "meta" ]; then
			stop_meta
		elif [ "$role" = "master" ]; then
			stop_master
		elif [ "$role" = "slave" ]; then
			stop_slave
		elif [ "$role" = "sentinel" ]; then
			stop_sentinel
		else
			usage
		fi
		;;
	C*)
		usage
		;;
esac

