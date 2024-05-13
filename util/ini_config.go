package util

import (
	"gopkg.in/ini.v1"
	"strconv"
)

// IniParser ini
type IniParser struct {
	confReader     *ini.File // config reader
	configFileName string
}

// IniParserError error
type IniParserError struct {
	errorInfo string
}

// Error
func (e *IniParserError) Error() string { return e.errorInfo }

// Load 加载
func (_self *IniParser) Load(configFileName string) error {
	conf, err := ini.Load(configFileName)
	if err != nil {
		_self.confReader = nil
		return err
	}
	_self.confReader = conf
	_self.configFileName = configFileName
	return nil
}

// GetString 获取字符串
func (_self *IniParser) GetString(section string, key string, defaultValue string) string {
	if _self.confReader == nil {
		return ""
	}

	s := _self.confReader.Section(section)
	if s == nil || s.Key(key).String() == "" {
		return defaultValue
	}

	return s.Key(key).String()
}

// GetInt32 int32
func (_self *IniParser) GetInt32(section string, key string, defaultValue int32) int32 {
	if _self.confReader == nil {
		return defaultValue
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return defaultValue
	}
	valueInt, _ := s.Key(key).Int()
	if valueInt == 0 {
		return defaultValue
	}

	return int32(valueInt)
}

// GetUint32 uint32
func (_self *IniParser) GetUint32(section string, key string, defaultValue uint32) uint32 {
	if _self.confReader == nil {
		return defaultValue
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return defaultValue
	}

	valueInt, _ := s.Key(key).Uint()
	if 0 == valueInt {
		return defaultValue
	}
	return uint32(valueInt)
}

// GetInt64 int64
func (_self *IniParser) GetInt64(section string, key string, defaultValue int64) int64 {
	if _self.confReader == nil {
		return defaultValue
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return defaultValue
	}

	valueInt, _ := s.Key(key).Int64()
	if valueInt == 0 {
		return defaultValue
	}
	return valueInt
}

// GetUint64 uint64
func (_self *IniParser) GetUint64(section string, key string, defaultValue uint64) uint64 {
	if _self.confReader == nil {
		return defaultValue
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return defaultValue
	}

	valueInt, _ := s.Key(key).Uint64()
	if valueInt == 0 {
		return defaultValue
	}
	return valueInt
}

// GetFloat32 float32
func (_self *IniParser) GetFloat32(section string, key string, defaultValue float32) float32 {
	if _self.confReader == nil {
		return defaultValue
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return defaultValue
	}

	valueFloat, _ := s.Key(key).Float64()
	if valueFloat == 0 {
		return defaultValue
	}
	return float32(valueFloat)
}

// GetFloat64 float64
func (_self *IniParser) GetFloat64(section string, key string, defaultValue float64) float64 {
	if _self.confReader == nil {
		return defaultValue
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return defaultValue
	}

	valueFloat, _ := s.Key(key).Float64()
	if valueFloat == 0 {
		return defaultValue
	}
	return valueFloat
}

// SetString string
func (_self *IniParser) SetString(section string, key string, value string) bool {
	if _self.confReader == nil {
		return false
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return false
	}
	k := s.Key(key)
	if k == nil {
		return false
	}
	k.SetValue(value)
	return true

}

// SetUInt uint32
func (_self *IniParser) SetUInt(section string, key string, value uint32) bool {
	if _self.confReader == nil {
		return false
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return false
	}
	k := s.Key(key)
	if k == nil {
		return false
	}
	strValue := strconv.FormatInt(int64(value), 10)
	k.SetValue(strValue)
	return true

}

// SetInt int32
func (_self *IniParser) SetInt(section string, key string, value int32) bool {
	if _self.confReader == nil {
		return false
	}

	s := _self.confReader.Section(section)
	if s == nil {
		return false
	}
	k := s.Key(key)
	if k == nil {
		return false
	}
	strValue := strconv.FormatInt(int64(value), 10)
	k.SetValue(strValue)
	return true

}

// Save save
func (_self *IniParser) Save() {
	_self.confReader.SaveTo(_self.configFileName)
}

// 某个节点所有key
func (_self *IniParser) GetSectionAllKeyName(sectionName string) []string {
	return _self.confReader.Section(sectionName).KeyStrings()
}

// 某个节点所有key,val
func (_self *IniParser) GetSectionData(sectionName string) map[string]string {
	return _self.confReader.Section(sectionName).KeysHash()
}
