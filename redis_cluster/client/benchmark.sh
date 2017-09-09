#!/usr/bin/env bash
# ******************************************************
# DESC    : test redis with tool redis-benchmark
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-04-10 00:34
# FILE    : benchmark.sh
# ******************************************************

./bin/redis-benchmark -p $1 -n 900 -k 1 -d 29 -q
