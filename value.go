package stationip

import (
	"github.com/yangchenxing/go-regionid"
)

// NetType 网络类型
type NetType int

const (
	// UnknownNetType 未知网络类型
	UnknownNetType NetType = iota

	// Station 基站
	Station

	// WiFi Wi-Fi
	WiFi
)

var (
	netTypeName = map[NetType]string{
		UnknownNetType: "Unknown",
		Station:        "Station",
		WiFi:           "Wi-Fi",
	}
)

// String 返回网络类型名称
func (nt NetType) String() string {
	return netTypeName[nt]
}

// Result 搜索结果
type Result struct {
	// 网络类型
	NetType NetType

	// 定位城市
	Location regionid.Location

	// 网络拥有者
	Owner string

	// 网络运营商
	Operator string
}

var (
	// NotFound 未找到时的返回结果
	NotFound = Result{
		NetType:  UnknownNetType,
		Location: nil,
		Owner:    "",
		Operator: "",
	}
)

// Equal 对比两个Result是否相等
func (v Result) Equal(other interface{}) bool {
	w, ok := other.(*Result)
	if !ok {
		return false
	}
	return v.NetType == w.NetType && v.Location == w.Location && v.Owner == w.Owner && v.Operator == w.Operator
}

func parseResult(fields []string) Result {
	location := regionid.GetLocation(fields[0], fields[1], fields[2])
	owner := fields[3]
	operator := fields[4]
	if owner == "*" {
		owner = operator
	}
	var netType NetType
	switch fields[5] {
	case "WIFI":
		netType = WiFi
	case "基站":
		netType = Station
	default:
		netType = UnknownNetType
	}
	return Result{
		NetType:  netType,
		Location: location,
		Owner:    owner,
		Operator: operator,
	}
}
