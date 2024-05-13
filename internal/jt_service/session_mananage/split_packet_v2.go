package jt_session

type splitPacketContextV2 struct {
	splitPacketContextBase
}

func (_self *splitPacketContextV2) initCodec(isPrintHex bool, M1, IA1, IC1, key uint32, cbNetMsgPacketFunc OnNetMsgPacketFunc) bool {
	return true
}

func (_self *splitPacketContextV2) putData(data []byte) bool {
	return _self.writeDataAndParse(data)
}
