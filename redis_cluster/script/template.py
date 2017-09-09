#!/bin/python
# -*- coding: UTF-8 -*-
# ******************************************************
# DESC    :
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-04-11 20:15
# FILE    : template.py
# ******************************************************

import os, sys

master_run_script_fmt="""#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:redis-master devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-03-30 15:29
# FILE    : meta-master-load.sh
# ******************************************************

# meta-master-script-name {start|stop} meta-master-ip meta-master-port
# user-param: {start|stop} meta-master-ip meta-master-port
sh bin/master-run.sh $1 %s $2 $3"""

slave_run_script_fmt="""#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:redis-slave devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-03-30 15:29
# FILE    : slave-load.sh
# ******************************************************

# slave-script-name {start|stop} index slave-ip slave-port master-ip master-port
# user-param: {start|stop} slave-ip slave-port master-ip master-port
sh bin/slave-run.sh $1 %s $2 $3 $4 $5"""

meta_master_run_script_fmt="""#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:meta-master devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-03-30 15:29
# FILE    : meta-master-load.sh
# ******************************************************

# meta-master-script-name {start|stop} meta-master-ip meta-master-port
# user-param: {start|stop} meta-master-ip meta-master-port
sh bin/meta-master-run.sh $1 $2 $3"""

meta_slave_run_script_fmt="""#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:meta-slave devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-03-30 15:29
# FILE    : meta-slave-load.sh
# ******************************************************

# meta-slave-script-name {start|stop} meta-slave-ip meta-slave-port meta-master-ip meta-master-port
# user-param: {start|stop} meta-slave-ip meta-slave-port meta-master-ip meta-master-port
sh bin/meta-slave-run.sh $1 $2 $3 $4 $5"""

sentinel_run_script_fmt="""#!/usr/bin/env bash
# ******************************************************
# DESC    : redis-cluster:redis-sentinel devops script
# AUTHOR  : Alex Stocks
# VERSION : 1.0
# LICENCE : LGPL V3
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-03-31 11:38
# FILE    : sentinel-load.sh
# ******************************************************

# sentinel-script-name {start|stop} index sentinel-ip sentinel-port notify-script instance-set
# user-param: {start|stop} sentinel-ip sentinel-port notify-script instance-script
sh bin/sentinel-run.sh $1 %s $2 $3 $4 $5"""

"""print help"""
def printHelp():
    """ print help prompt
    """
    print 'usage:'
    print '  example: ./template.py {master|slave} index memory-size'
    print '  example: ./template.py {meta-master|meta-slave} memory-size'
    print '  example: ./template.py sentinel index'

def saveFile(filename, contents):
  fh = open(filename, 'w+')
  fh.write(contents)
  fh.close()

def genMaster(index, memory_size):
    content = (master_run_script_fmt % (index))
    saveFile('master-load.sh', content)
    dir = 'master%s' % index
    cmd = ("mkdir -p %s && cd %s && mv ../master-load.sh ./ && "
            "mkdir -p bin && cp ../template/redis-server ./bin/redis-server%s && "
            "cp ../template/master-run.sh ./bin/ && "
            "sed -i \"s/maxmemory=1G/maxmemory=%s/g\" bin/master-run.sh &&"
            "mkdir -p conf && cp ../template/redis.conf.template conf/"
            % (dir, dir, index, memory_size))
    # print cmd
    os.system(cmd)

def genSlave(index, memory_size):
    content = (slave_run_script_fmt % (index))
    saveFile('slave-load.sh', content)
    dir = 'slave%s' % index
    cmd = ("mkdir -p %s && cd %s && mv ../slave-load.sh ./ && "
            "mkdir -p bin && cp ../template/redis-server ./bin/redis-slave%s && "
            "cp ../template/slave-run.sh ./bin/ && "
            "sed -i \"s/maxmemory=1G/maxmemory=%s/g\" bin/slave-run.sh &&"
            "mkdir -p conf && cp ../template/redis.conf.template conf/"
            % (dir, dir, index, memory_size))
    # print cmd
    os.system(cmd)

def genMetaMaster(memory_size):
    content = (meta_master_run_script_fmt)
    saveFile('meta-master-load.sh', content)
    dir = 'meta_master'
    cmd = ("mkdir -p %s && cd %s && mv ../meta-master-load.sh ./ && "
            "mkdir -p bin && cp ../template/redis-server ./bin/meta-redis-server && "
            "cp ../template/meta-master-run.sh ./bin/ && "
            "sed -i \"s/maxmemory=1G/maxmemory=%s/g\" bin/meta-master-run.sh &&"
            "mkdir -p conf && cp ../template/redis.conf.template conf/"
            % (dir, dir, memory_size))
    # print cmd
    os.system(cmd)

def genMetaSlave(memory_size):
    content = (meta_slave_run_script_fmt)
    saveFile('meta-slave-load.sh', content)
    dir = 'meta_slave'
    cmd = ("mkdir -p %s && cd %s && mv ../meta-slave-load.sh ./ && "
            "mkdir -p bin && cp ../template/redis-server ./bin/meta-redis-slave && "
            "cp ../template/meta-slave-run.sh ./bin/ && "
            "sed -i \"s/maxmemory=1G/maxmemory=%s/g\" bin/meta-slave-run.sh &&"
            "mkdir -p conf && cp ../template/redis.conf.template conf/"
            % (dir, dir, memory_size))
    # print cmd
    os.system(cmd)

def genSentinel(index):
    content = (sentinel_run_script_fmt % (index))
    saveFile('sentinel-load.sh', content)
    dir = 'sentinel%s' % index
    cmd = ("mkdir -p %s && cd %s && mv ../sentinel-load.sh ./ && "
            "mkdir -p bin && cp ../template/redis-server ./bin/redis-sentinel%s && "
            "cp ../template/sentinel-run.sh ./bin/ && "
            "mkdir -p conf && cp ../template/sentinel.conf.template conf/"
            % (dir, dir, index))
    # print cmd
    os.system(cmd)

if __name__ == '__main__':
    if len(sys.argv) < 3:
        printHelp()
        sys.exit(1)

    role = sys.argv[1]
    index = sys.argv[2]
    if role == 'master':
        memory_size = sys.argv[3]
        genMaster(index, memory_size)
    elif role == 'slave':
        memory_size = sys.argv[3]
        genSlave(index, memory_size)
    elif role == 'meta-master':
        memory_size = sys.argv[2]
        genMetaMaster(memory_size)
    elif role == 'meta-slave':
        memory_size = sys.argv[2]
        genMetaSlave(memory_size)
    elif role == 'sentinel':
        genSentinel(index)
    else:
        printHelp()
        sys.exit(1)

