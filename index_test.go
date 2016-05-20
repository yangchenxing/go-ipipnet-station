package stationip

import (
	"bytes"
	"github.com/yangchenxing/go-regionid"
	"net"
	"strings"
	"testing"
)

var (
	emtpyIndex [256][2]int
)

func init() {
	for i := 0; i < 256; i++ {
		emtpyIndex[i][0], emtpyIndex[i][1] = -1, -1
	}
	regionid.LoadBuildWorld()
}

func TestIndex(t *testing.T) {
	data := []byte(strings.Join([]string{
		"014.024.121.000\t014.024.121.255\t中国\t广东\t*\t*\t电信\t基站",
		"014.024.122.000\t014.024.122.255\t中国\t广东\t*\t*\t电信\t基站",
		"014.024.123.000\t014.024.123.255\t中国\t广东\t*\t*\t电信\t基站",
		"014.024.124.000\t014.024.124.255\t中国\t广东\t*\t*\t电信\t基站",
		"014.024.125.000\t014.024.125.255\t中国\t广东\t*\t*\t电信\t基站",
		"027.149.109.000\t027.149.109.255\t中国\t福建\t*\t*\t电信\t基站",
		"027.149.110.000\t027.149.110.255\t中国\t福建\t*\t*\t电信\t基站",
		"027.149.111.000\t027.149.111.255\t中国\t福建\t*\t*\t电信\t基站",
		"027.149.112.000\t027.149.112.255\t中国\t福建\t*\t*\t电信\t基站",
		"027.149.113.000\t027.149.113.255\t中国\t福建\t*\t*\t电信\t基站",
	}, "\n"))
	actual, err := newIndex(bytes.NewBuffer(data), "")
	if err != nil {
		t.Error("new index fail:", err.Error())
		return
	}
	expected := &index{
		Sections: []section{
			{
				Lower: 0x0e187900,
				Upper: 0x0e187dFF,
				Result: Result{
					NetType:  Station,
					Location: regionid.GetLocation("中国", "广东", ""),
					Owner:    "电信",
					Operator: "电信",
				},
			},
			{
				Lower: 0x1b956d00,
				Upper: 0x1b9571FF,
				Result: Result{
					NetType:  Station,
					Location: regionid.GetLocation("中国", "福建", ""),
					Owner:    "电信",
					Operator: "电信",
				},
			},
		},
		Index: emtpyIndex,
		ETag:  "",
	}
	expected.Index[14][0], expected.Index[14][1] = 0, 0
	expected.Index[27][0], expected.Index[27][1] = 1, 1
	if !compareIndex(actual, expected) {
		t.Error("unexpected sections:", actual.Sections)
		return
	}
	if !compareResult(t, "14.24.121.1", actual.search(net.ParseIP("14.24.121.1")), Station, "广东", "电信") {
		return
	}
	if !compareResult(t, "14.24.129.1", actual.search(net.ParseIP("14.24.129.1")), UnknownNetType, "", "") {
		return
	}
}

func TestLocal(t *testing.T) {
	cfg := Config{
		LocalPath: "StationGWList.sample.txt",
	}
	if err := Initialize(cfg); err != nil {
		t.Error("intialize fail:", err.Error())
		return
	}
	if !compareResult(t, "1.50.5.1", MustSearch(net.ParseIP("1.50.5.1")), Station, "宁夏", "电信") {
		return
	}
	if !compareResult(t, "1.200.23.2", MustSearch(net.ParseIP("1.200.23.2")), Station, "台湾", "tstartel.com") {
		return
	}
}

func compareResult(t *testing.T, ip string, actual Result, netType NetType, province, operator string) bool {
	if actual.NetType != netType || actual.Operator != operator {
		t.Errorf("search ip %q fail: %v", ip, actual)
		return false
	}
	if province != "" && (actual.Location == nil || actual.Location.Subdivision() == nil || actual.Location.Subdivision().Name != province) {
		t.Errorf("search ip %q fail: %v", ip, actual)
		return false
	}
	return true
}

func compareIndex(a, b *index) bool {
	if len(a.Sections) != len(b.Sections) {
		return false
	}
	for i, x := range a.Sections {
		if b.Sections[i] != x {
			return false
		}
	}
	for i, x := range a.Index {
		if b.Index[i] != x {
			return false
		}
	}
	return a.ETag == b.ETag
}
