package jt_session

type splitPacketContextV1 struct {
	splitPacketContextBase
}

func (_self *splitPacketContextV1) initCodec(isPrintHex bool, M1, IA1, IC1, key uint32, cbNetMsgPacketFunc OnNetMsgPacketFunc) bool {
	return true
}
func (_self *splitPacketContextV1) putData(data []byte) bool {
	return _self.writeDataAndParse(data)
}
