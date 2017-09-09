#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:redis-master devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : zhaoxin@zenmen.com
# MOD     : 2016-03-30 15:29
# FILE    : master-run.sh
# ******************************************************

host=0.0.0.0
port=0
name="redis-server"
index=""

database_num=16
maxmemory=1G

start() {
	stop
	# cp -rf ./bin/$name ./bin/$name$index
	[[ -f conf/redis.conf ]] && cp conf/redis.conf conf/redis.conf.`date +%Y-%m-%d-%H-%M-%S`
	cp conf/redis.conf.template conf/redis.conf
	mkdir -p ./data/ ./log/ ./pid/
	./bin/$name$index conf/redis.conf --daemonize yes --bind $host --port $port --databases $database_num --maxmemory $maxmemory --maxmemory-policy volatile-lru --save 300 10 --save 900 1 --save 3600 1 --maxmemory-samples 3 --appendonly no --no-appendfsync-on-rewrite yes --slowlog-log-slower-than 10000 --slowlog-max-len 256 --list-max-ziplist-entries 512 --list-max-ziplist-value 64 --set-max-intset-entries 512 --zset-max-ziplist-entries 128 --zset-max-ziplist-value 64 --activerehashing yes --pidfile $(pwd)/pid/redis.pid --logfile $(pwd)/log/server.log --dir $(pwd)/data --dbfilename dump$index.rdb
	PID=`ps aux | grep -w $name$index | grep -v grep | awk '{print $2}'`
	if [ "$PID" != "" ];
	then
		echo "start $name$index ( pid =" $PID ")"
	fi
}

stop() {
	#killall -9 $name$index
	# ps aux | grep -w $name$index | grep -v grep | awk '{print $2}' | xargs kill -9
	PID=`ps aux | grep -w $name$index | grep -v grep | awk '{print $2}'`
	if [ "$PID" != "" ];
	then
		echo "kill -9 $name$index ( pid =" $PID ")"
		kill -9 $PID
	fi
}

clean() {
	mv conf/redis.conf.template ./ && rm -rf ./data/* ./log/* ./pid/* ./conf/* && mv redis.conf.template ./conf/
}

case C"$1" in
	C)
		echo "Usage: $0 {start|stop}"
		;;
	Cstart)
		if [ $# != 4 ]; then
			echo "Please Input: index host port"
		else
			index=$2
			host=$3
			port=$4
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

