package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

import (
	"github.com/AlexStocks/goext/database/redis"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

type (
	ClusterMeta struct {
		Version   int32                       `json:"version, omitempty"`
		Instances map[string]gxredis.Instance `json:"instances, omitempty"`
	}
	InstanceNameList struct {
		list []string `json:"list, omitempty"`
	}
	SentinelWorker struct {
		sntl *gxredis.Sentinel
		// redis instances meta data
		sync.RWMutex
		meta    ClusterMeta
		wg      sync.WaitGroup
		watcher *gxredis.SentinelWatcher
	}
)

func NewSentinelWorker() *SentinelWorker {
	var (
		err       error
		res       interface{}
		instances []gxredis.Instance
		metaDB    gxredis.Instance
		metaConn  redis.Conn
		version   int
	)

	worker := &SentinelWorker{
		sntl: gxredis.NewSentinel(Conf.Redis.Sentinels),
		meta: ClusterMeta{
			Instances: make(map[string]gxredis.Instance, 32),
		},
	}

	instances, err = worker.sntl.GetInstances()
	if err != nil {
		panic(fmt.Sprintf("st.GetInstances, error:%#v\n", err))
	}

	for _, inst := range instances {
		if inst.Name == Conf.Redis.MetaDBName {
			metaDB = inst
		}
		// discover new sentinel
		err = worker.sntl.Discover(inst.Name)
		if err != nil {
			panic(fmt.Sprintf("failed to discover sentiinels of instance:%s, error:%#v", inst.Name, err))
		}
	}

	worker.meta.Version = 1
	if metaDB.Name == "" {
		panic("can not find meta db.")
	}
	if metaConn, err = worker.sntl.GetConnByRole(metaDB.Master.String(), gxredis.RR_Master); err != nil {
		panic(fmt.Sprintf("gxsentinel.GetConnByRole(%s, RR_Master) = %#v", metaDB.Master.String(), err))
	}
	defer metaConn.Close()
	if res, err = metaConn.Do("hget", Conf.Redis.MetaHashtable, Conf.Redis.MetaVersion); err != nil {
		panic(fmt.Sprintf("hget(%s, %s, %s) = error:%#v", Conf.Redis.MetaHashtable, Conf.Redis.MetaVersion, err))
	}
	if version, err = strconv.Atoi(string(res.([]byte))); err != nil {
		panic(fmt.Sprintf("strconv.Atoi(%#v) = error:%#v", res, err))
	}
	worker.meta.Version = int32(version)

	worker.updateClusterMeta()

	return worker
}

func (w *SentinelWorker) storeClusterMetaData() error {
	var (
		err              error
		ok               bool
		jsonStr          []byte
		metaDB           gxredis.Instance
		metaConn         redis.Conn
		instanceNameList InstanceNameList
	)

	w.RLock()
	defer w.RUnlock()

	if len(w.meta.Instances) == 0 {
		return fmt.Errorf("redis cluster instance pool is empty.")
	}

	if metaDB, ok = w.meta.Instances[Conf.Redis.MetaDBName]; !ok {
		return fmt.Errorf("can not find meta db.")
	}

	if metaConn, err = w.sntl.GetConnByRole(metaDB.Master.String(), gxredis.RR_Master); err != nil {
		return errors.Wrapf(err, "gxsentinel.GetConnByRole(%s, RR_Master)", metaDB.Master.String())
	}
	defer metaConn.Close()

	if _, err = metaConn.Do("hset", Conf.Redis.MetaHashtable, Conf.Redis.MetaVersion, w.meta.Version); err != nil {
		return errors.Wrapf(err, "hset(%s, %s, %s)", Conf.Redis.MetaHashtable, Conf.Redis.MetaVersion, w.meta.Version)
	}
	for k, v := range w.meta.Instances {
		if k != Conf.Redis.MetaDBName {
			if jsonStr, err = json.Marshal(v); err != nil {
				Log.Error("json.Marshal(%#v) = %#v", v, err)
				continue
			}
			if _, err = metaConn.Do("hset", Conf.Redis.MetaHashtable, k, string(jsonStr)); err != nil {
				Log.Error(err, "hset(%s, %s, %s) = error:%#v", Conf.Redis.MetaHashtable, k, string(jsonStr), err)
				continue
			}
			instanceNameList.list = append(instanceNameList.list, k)
		}
	}
	if jsonStr, err = json.Marshal(instanceNameList); err != nil {
		return errors.Wrapf(err, "json.Marshal(%#v)", instanceNameList)
	}
	if _, err = metaConn.Do("hset", Conf.Redis.MetaHashtable, Conf.Redis.MetaInstNameList, string(jsonStr)); err != nil {
		return errors.Wrapf(err, "hset(%s, %s, %s)", Conf.Redis.MetaHashtable, Conf.Redis.MetaInstNameList, string(jsonStr))
	}

	return nil
}

func (w *SentinelWorker) updateClusterMeta() error {
	instances, err := w.sntl.GetInstances()
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("st.GetInstances, error:%#v\n", err))
	}

	var flag bool
	for _, inst := range instances {
		// discover new sentinel
		err = worker.sntl.Discover(inst.Name)
		if err != nil {
			return errors.Wrapf(err, "failed to discover sentiinels of instance:%s, error:%#v", inst.Name, err)
		}
		slaves := inst.Slaves
		// delete unavailable slave
		inst.Slaves = inst.Slaves[:0]
		for _, slave := range slaves {
			if slave.Available() {
				inst.Slaves = append(inst.Slaves, slave)
			}
		}

		w.RLock()
		redisInst, ok := w.meta.Instances[inst.Name]
		w.RUnlock()
		if ok { // 在原来name已经存在的情况下，再查验instance值是否相等
			instJson, _ := json.Marshal(inst)
			redisInstJson, _ := json.Marshal(redisInst)
			ok = string(instJson) == string(redisInstJson)
		}
		if !ok {
			w.Lock()
			flag = true
			w.meta.Instances[inst.Name] = inst
			w.Unlock()
		}
	}
	if flag {
		w.Lock()
		w.meta.Version++
		w.Unlock()
		// update meta data to meta redis
		if err = w.storeClusterMetaData(); err != nil {
			return errors.Wrapf(err, "SentinelWorker.storeClusterMetaData()")
		}
	}

	return nil
}

func (w *SentinelWorker) updateClusterMetaByInstanceSwitch(info gxredis.MasterSwitchInfo) {
	w.Lock()
	defer w.Unlock()
	inst, ok := w.meta.Instances[info.Name]
	inst.Name = info.Name
	inst.Master = info.NewMaster
	if !ok {
		w.meta.Instances[inst.Name] = inst
		w.meta.Version++
		return
	}

	for i, slave := range inst.Slaves {
		if string(slave.IP) == string(info.NewMaster.IP) && slave.Port == info.NewMaster.Port {
			inst.Slaves = append(inst.Slaves[:i], inst.Slaves[i+1:]...)
			break
		}
	}
	w.meta.Version++
}

func (w *SentinelWorker) WatchInstanceSwitch() error {
	var (
		err error
	)
	w.watcher, err = w.sntl.MakeSentinelWatcher()
	if err != nil {
		return errors.Wrapf(err, "MakeSentinelWatcher")
	}
	c, _ := w.watcher.Watch()
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for addr := range c {
			Log.Info("redis instance switch info: %#v\n", addr)
			w.updateClusterMetaByInstanceSwitch(addr)
			w.storeClusterMetaData()
		}
		Log.Info("instance switch watch exit")
	}()

	return nil
}

func (w *SentinelWorker) Close() {
	w.watcher.Close()
	w.wg.Wait()
	w.sntl.Close()
}
