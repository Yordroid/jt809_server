package data_manage_service

import (
	"jt809_server/models"
	"jt809_server/util"
	"sync"
)

//用来缓存车辆的基础数据

type dataManageContext struct {
	util.GoService
	rwMutex               sync.RWMutex
	mapVehicleInfo        map[uint32]*models.VsVehicleBaseInfo
	mapVehicleNoVehicleID map[string]uint32
	mapVehicleUserKey     map[uint32]*util.Set[int64]
}

func (_self *dataManageContext) initService() {
	_self.mapVehicleInfo = make(map[uint32]*models.VsVehicleBaseInfo)
	_self.mapVehicleNoVehicleID = make(map[string]uint32)
	_self.mapVehicleUserKey = make(map[uint32]*util.Set[int64])
	_self.SetServiceStartedFinish()
}

func (_self *dataManageContext) updateUserVehicleMapInfo(userKey int64, vehicleIDs []uint32) {
	//先清除用户关联的车辆
	_self.clearUserVehicleMapInfo(userKey)
	_self.rwMutex.Lock()
	for idx := 0; idx < len(vehicleIDs); idx++ {
		userKeySet, isExist := _self.mapVehicleUserKey[vehicleIDs[idx]]
		if !isExist {
			userKeySet = &util.Set[int64]{}
			_self.mapVehicleUserKey[vehicleIDs[idx]] = userKeySet
		}
		userKeySet.Add(userKey)
	}

	_self.rwMutex.Unlock()
}
func (_self *dataManageContext) clearUserVehicleMapInfo(userKey int64) {
	_self.rwMutex.Lock()
	defer _self.rwMutex.Unlock()
	for _, userKeySet := range _self.mapVehicleUserKey {
		var delKeys []int64
		userKeySet.ForEach(func(userKeyQuery int64) {
			if userKeyQuery == userKey {
				delKeys = append(delKeys, userKey)
			}
		})
		for idx := 0; idx < len(delKeys); idx++ {
			userKeySet.Del(delKeys[idx])
		}
	}
}

func (_self *dataManageContext) foreachUserVehicleByVehicleID(vehicleID uint32, userKeyFunc func(int64)) {
	var userKeys []int64
	_self.rwMutex.RLock()
	userKeySet, isExist := _self.mapVehicleUserKey[vehicleID]
	if isExist {
		userKeySet.ForEach(func(userKey int64) {
			userKeys = append(userKeys, userKey)
		})
	}
	_self.rwMutex.RUnlock()
	for idx := 0; idx < len(userKeys); idx++ {
		if userKeyFunc != nil {
			userKeyFunc(userKeys[idx])
		}
	}
}

func (_self *dataManageContext) getVehicleInfoByID(vehicleID uint32) *models.VsVehicleBaseInfo {
	_self.rwMutex.RLock()
	defer _self.rwMutex.RUnlock()
	vehicleInfo, isExist := _self.mapVehicleInfo[vehicleID]
	if isExist {
		return vehicleInfo
	}
	return nil
}

func (_self *dataManageContext) getVehicleInfoByVehicleNo(vehicleNo string) *models.VsVehicleBaseInfo {
	_self.rwMutex.RLock()
	defer _self.rwMutex.RUnlock()
	vehicleID, isExist := _self.mapVehicleNoVehicleID[vehicleNo]
	if isExist {
		vehicleInfo, isExist := _self.mapVehicleInfo[vehicleID]
		if isExist {
			return vehicleInfo
		}
	}
	return nil
}

func (_self *dataManageContext) updateVehicleInfo(vehicleInfo *models.VsVehicleBaseInfo) {
	_self.rwMutex.Lock()
	defer _self.rwMutex.Unlock()
	_self.mapVehicleInfo[vehicleInfo.VehicleID] = vehicleInfo
	_self.mapVehicleNoVehicleID[vehicleInfo.VehicleNo] = vehicleInfo.VehicleID
}

func (_self *dataManageContext) getAllVehicleNum(userKey int64) int {
	_self.rwMutex.RLock()
	defer _self.rwMutex.RUnlock()
	vehicleNum := 0
	for _, userKeySet := range _self.mapVehicleUserKey {
		userKeySet.ForEach(func(queryUserKey int64) {
			if queryUserKey == userKey {
				vehicleNum++
			}
		})
	}
	return vehicleNum
}
