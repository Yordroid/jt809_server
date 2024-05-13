package data_manage_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/models"
	"jt809_server/util"
	"sync"
)

// DataManageApi 数据管理服务对外接口,单例

type DataManageApi struct {
	ctx *dataManageContext
}

var api *DataManageApi
var apiOnce sync.Once

// DataManageServiceIns 实例
func DataManageServiceIns() *DataManageApi {
	apiOnce.Do(func() {
		api = &DataManageApi{}
	})
	return api
}

func (_self *DataManageApi) InitApi(notifyFrameWork *util.IotMsgNotifyFramework) {
	_self.ctx = &dataManageContext{}
	_self.ctx.StartService("dataManageContext", 1000, notifyFrameWork, func() {
		log.Info("dataManageContext start finish")
		_self.ctx.initService()
	})
}

// GetVehicleInfoByID 获取车辆信息,支持并发读
func (_self *DataManageApi) GetVehicleInfoByID(vehicleID uint32) *models.VsVehicleBaseInfo {
	return _self.ctx.getVehicleInfoByID(vehicleID)
}

// GetVehicleInfoByVehicleNo 获取车辆信息,支持并发读
func (_self *DataManageApi) GetVehicleInfoByVehicleNo(vehicleNo string) *models.VsVehicleBaseInfo {
	return _self.ctx.getVehicleInfoByVehicleNo(vehicleNo)
}

// UpdateVehicleInfo 更新车辆信息
func (_self *DataManageApi) UpdateVehicleInfo(vehicleInfo *models.VsVehicleBaseInfo) {
	_self.ctx.updateVehicleInfo(vehicleInfo)
}

// GetAllVehicleNum 获取用户所有车辆的数量
func (_self *DataManageApi) GetAllVehicleNum(userKey int64) int {
	return _self.ctx.getAllVehicleNum(userKey)
}

// UpdateUserVehicleMapInfo 更新用户和车辆的映射关系
func (_self *DataManageApi) UpdateUserVehicleMapInfo(userKey int64, vehicleIDs []uint32) {
	_self.ctx.updateUserVehicleMapInfo(userKey, vehicleIDs)
}

// DeleteUserVehicleMapInfo 删除用户和车辆的映射关系
func (_self *DataManageApi) DeleteUserVehicleMapInfo(userKey int64) {
	_self.ctx.clearUserVehicleMapInfo(userKey)
}

// ForeachUserVehicleByVehicleID 遍历用户信息通过车辆ID
func (_self *DataManageApi) ForeachUserVehicleByVehicleID(vehicleID uint32, userKeyFunc func(int64)) {
	_self.ctx.foreachUserVehicleByVehicleID(vehicleID, userKeyFunc)
}
