package stationip

import (
	"github.com/yangchenxing/go-ipipnet-downloader"
	"net"
	"testing"
)

func TestSearch(t *testing.T) {
	index := &Index{
		Downloader: downloader.Downloader{
			LocalPath: "StationGWList.sample.txt",
		},
	}
	if err := index.Initialize(); err != nil {
		t.Error("initialize index fail:", err.Error())
		t.FailNow()
	}
	// 001.050.014.001
	result, err := index.Search(net.ParseIP("001.050.014.001"))
	if err != nil {
		t.Error("search 001.050.014.001 fail:", err.Error())
		t.FailNow()
	}
	if result.NetType != Station || result.Location.Subdivision().Name() != "宁夏" || result.Operator != "电信" {
		t.Error("unexpected result of 001.050.014.001:", result)
		t.FailNow()
	}
	// 001.087.139.001
	result, err = index.Search(net.ParseIP("001.087.139.001"))
	if err != nil {
		t.Error("search 001.087.139.001 fail:", err.Error())
		t.FailNow()
	}
	if result.NetType != WiFi || result.Location.City().Name() != "安康" || result.Operator != "电信" {
		t.Error("unexpected result of 001.087.139.001:", result)
		t.FailNow()
	}
	// 254.254.254.254
	result, err = index.Search(net.ParseIP("254.254.254.254"))
	if err != nil {
		t.Error("search 254.254.254.254 fail:", err.Error())
		t.FailNow()
	}
	if result != NotFound {
		t.Error("unexpected result of 254.254.254.254:", result)
		t.FailNow()
	}
}
