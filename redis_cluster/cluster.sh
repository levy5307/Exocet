#!/usr/bin/env bash
# ******************************************************
# DESC    : PalmChat redis-cluster cluster template script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-04-06 16:41
# FILE    : cluster.sh
# ******************************************************

meta_db_mem_size='50m'
local_ip=0.0.0.0

instance_name=cache
master_port=4000
slave_port=14000

meta_name=meta
meta_master_port=6000
meta_slave_port=6001

gen() {
	cd script
	# gen redis instances
	for ((idx = 0; idx < $1; idx ++))
	do
		python template.py master $idx $2
		python template.py slave $idx $2
	done

	# gen metadb
	python template.py meta-master $meta_db_mem_size
	python template.py meta-slave $meta_db_mem_size

	# gen sentinel isntances
	for ((idx = 0; idx < 3; idx ++))
	do
		python template.py sentinel $idx
	done

	rm -rf ./cluster
	mkdir -p ./cluster
	mv master* slave* sentinel* meta* ./cluster/
	mkdir -p ./cluster/script/
	cp cluster_notify.sh ./cluster/script/
	cp cluster_notify.sh /tmp/
	cp ../cluster.sh cluster/
	cp -rf ../client/ cluster/
	rm -rf ../cluster/
	mv cluster ../
}

start() {
	# local_ip=10.136.40.61
	instance_set=""
	# start redis instances
	for ((idx = 0; idx < $1; idx ++))
	do
		[[ -d master$idx ]] && cd master$idx && sh master-load.sh start $local_ip $master_port && cd ..
		sleep 2
		[[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh  start $local_ip $slave_port $local_ip $master_port && cd ..
		instance_set="${instance_set}@${instance_name}${idx}:$local_ip:${master_port}"
		((master_port ++))
		((slave_port ++))
	done

	# start meta redis
	[[ -d meta_master ]] && cd meta_master && sh meta-master-load.sh start $local_ip $meta_master_port && cd ..
	sleep 2
	[[ -d meta_slave ]] && cd meta_slave && sh meta-slave-load.sh start $local_ip $meta_slave_port $local_ip $meta_master_port && cd ..
	instance_set="${instance_set}@${meta_name}:$local_ip:${meta_master_port}"

	# start sentinels
	instance_set=${instance_set:1}
	cp script/cluster_notify.sh /tmp/
	sentinel_port=26380
	for ((idx = 0; idx < 3; idx ++))
	do
		[[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh start $local_ip $sentinel_port /tmp/cluster_notify.sh $instance_set && cd ..
		((sentinel_port ++))
	done
}

stop() {
	# stop redis instances
	for ((idx = 0; idx < $1; idx ++))
	do
		[[ -d master$idx ]] && cd master$idx && sh master-load.sh stop && cd ..
		sleep 2
		[[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh  stop && cd ..
	done
	sleep 2

	# stop meta redis instances
	[[ -d meta_master ]] && cd meta_master && sh meta-master-load.sh stop && cd ..
	[[ -d meta_slave ]] && cd meta_slave && sh meta-slave-load.sh stop && cd ..
	sleep 2

	# stop sentinels
	for ((idx = 0; idx < 3; idx ++))
	do
		[[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh stop && cd ..
	done
}

clean() {
	# clean redis instances
	for ((idx = 0; idx < $1; idx ++))
	do
		[[ -d master$idx ]] && cd master$idx && sh master-load.sh clean && cd ..
		[[ -d slave$idx ]] && cd slave$idx  && sh slave-load.sh  clean && cd ..
	done
	sleep 2

	# clean meta redis instances
	[[ -d meta_master ]] && cd meta_master && sh meta-master-load.sh clean && cd ..
	[[ -d meta_slave ]] && cd meta_slave && sh meta-slave-load.sh clean && cd ..
	sleep 2

	sleep 2
	# clean sentinels
	for ((idx = 0; idx < 3; idx ++))
	do
		[[ -d sentinel$idx ]] && cd sentinel$idx && sh sentinel-load.sh clean && cd ..
	done
}

case C"$1" in
	Cstart)
		if [ "-$2" = "-" ]; then
			echo "Please Input: $0 start instance-number"
		else
			start $2
			echo "start Done!"
		fi
		;;
	Cstop)
		if [ "-$2" = "-" ]; then
			echo "Please Input: $0 stop instance-number"
		else
			stop $2
			echo "stop Done!"
		fi
		;;
	Cgen)
		if [ "-$3" = "-" ]; then
			echo "Please Input: $0 gen instance-number memory-size"
		else
			gen $2 $3
			echo "gen Done!"
		fi
		;;
	Cclean)
		if [ "-$2" = "-" ]; then
			echo "Please Input: $0 clean instance-number"
		else
			stop $2
			clean $2
			echo "clean Done!"
		fi
		;;
	C*)
		echo "Usage: $0 {start|stop|clean|gen}"
		;;
esac

