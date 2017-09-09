#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:redis-sentinel devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-03-30 20:22
# FILE    : sentinel-run.sh
# ******************************************************

host=0.0.0.0
port=0
name="redis-sentinel"
index="0"
sentinel_str=""
sentinel_params=""
notify_script=""

parseNodeInfo() {
	host_name=""
	host_ip=""
	host_port=""
	if [[ $1 =~ ^([a-zA-Z0-9]{1,}):([0-9]{1,2}|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]{1,2}|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]{1,2}|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\.([0-9]{1,2}|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\:([0-9]{1,4}|[1-5][0-9]{4}|6[0-5]{2}[0-3][0-5])$ ]]
    then
        host_name=${BASH_REMATCH[1]}
        host_ip=${BASH_REMATCH[2]}.${BASH_REMATCH[3]}.${BASH_REMATCH[4]}.${BASH_REMATCH[5]}
        host_port=${BASH_REMATCH[6]}
    fi
	sentinel_str=""
    if [ -n "$host_name" -a -n "$host_ip" -a -n "$host_port" ]; then
		host_name="${host_name}"
		sentinel_str=" --sentinel monitor $host_name $host_ip $host_port 2 --sentinel down-after-milliseconds $host_name 15000 --sentinel parallel-syncs $host_name 1 --sentinel failover-timeout $host_name 450000 --sentinel client-reconfig-script $host_name $notify_script"
    fi
}

start() {
	stop
	# cp -rf ./bin/$name ./bin/$name$index
	[[ -f conf/sentinel.conf ]] && cp conf/sentinel.conf conf/sentinel.conf.`date +%Y-%m-%d-%H-%M-%S`
	cp conf/sentinel.conf.template conf/sentinel.conf
	mkdir -p ./data/ ./log/ ./pid/
	./bin/$name$index conf/sentinel.conf --daemonize yes --bind $host --port $port --pidfile $(pwd)/pid/redis.pid --logfile $(pwd)/log/server.log --dir $(pwd)/data $sentinel_params
	PID=`ps aux | grep -w $name$index | grep -v grep | awk '{print $2}'`
	if [ "$PID" != "" ];
	then
		echo "start $name$index ( pid =" $PID ")"
	fi
}

stop() {
	# killall -9 $name$index
	# pkill -9 $name$index
	# ps aux | grep -w $name$index | grep -v grep | awk '{print $2}' | xargs kill -9
	PID=`ps aux | grep -w $name$index | grep -v grep | awk '{print $2}'`
	if [ "$PID" != "" ];
	then
		echo "kill -9 $name$index ( pid =" $PID ")"
		kill -9 $PID
	fi
}

clean() {
	mv conf/sentinel.conf.template ./ && rm -rf ./data/* ./log/* ./pid/* ./conf/* && mv sentinel.conf.template ./conf/
}

case C"$1" in
	C)
		echo "Usage: $0 {start|stop}"
		;;

	Cstart)
		if [ "-$4" = "-" ]; then
			echo "Please Input: $0 start index host port redis-node-name master-ip:master-port"
		else
			index=$2
			host=$3
			port=$4
			notify_script=$5
			instance_set=$6
			instance_num=`echo $instance_set | grep -o '@' | wc -l`
			((instance_num++))
			idx_max=$instance_num
			for ((idx = 1; idx <= $idx_max; idx ++))
			do
				# echo $instance_set | awk -F '@' '{print $idx}'
				instance=`echo $instance_set | cut -d "@" -f$idx`
				parseNodeInfo $instance
				sentinel_params="${sentinel_params}${sentinel_str}"

			done
			# echo $sentinel_params
			start
			echo "Done!"
		fi
		;;

	Cstop)
		if [ "-$2" = "-" ]; then
			echo "Please Input: index"
		else
			index=$2
			stop
			echo "Done!"
		fi
		;;
	Cclean)
		clean
		;;
	C*)
		echo "Usage: $0 {start|stop}"
		;;
esac
