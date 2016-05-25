package stationip

import (
	"bufio"
	"fmt"
	"github.com/yangchenxing/go-ip-index"
	"github.com/yangchenxing/go-ipipnet-downloader"
	"github.com/yangchenxing/go-regionid"
	"net"
	"os"
	"strings"
)

// Index 基站IP库索引
type Index struct {
	// IPIP.net数据下载器
	downloader.Downloader

	// 索引库最小二分查找范围
	MinBinarySearchRange int

	index *ipindex.IPIndex
}

// Initialize 初始化Index
func (index *Index) Initialize() error {
	if !regionid.Initialized() {
		if err := regionid.LoadBuiltinWorld(); err != nil {
			return fmt.Errorf("load buildin regionid fail: %s", err.Error())
		}
	}
	if index.MinBinarySearchRange <= 0 {
		index.MinBinarySearchRange = ipindex.DefaultMinBinarySearchRange
	}
	if err := index.EnsureLocal(); err != nil {
		return fmt.Errorf("ensure local file fail: %s", err.Error())
	}
	if err := index.load(); err != nil {
		return fmt.Errorf("load local file fail: %s", err.Error())
	}
	index.UpdateCallback = func(string) { index.load() }
	go index.StartWatch()
	return nil
}

// Search 检索IP所属网络类型
func (index *Index) Search(ip net.IP) (result Result, err error) {
	var value ipindex.Value
	value, err = index.index.Search(ip)
	if err == nil && value != nil {
		result = value.(Result)
	}
	return
}

func (index *Index) load() error {
	builder := ipindex.NewIndexBuilder(index.MinBinarySearchRange)
	file, err := os.Open(index.LocalPath)
	if err != nil {
		return fmt.Errorf("open local file fail: %s", err.Error())
	}
	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		text := strings.TrimRight(scanner.Text(), "\n")
		if len(text) == 0 {
			continue
		}
		fields := strings.Split(text, "\t")
		if len(fields) != 8 {
			return fmt.Errorf("invalid line %d: %q", i, text)
		}
		err := builder.Add(net.ParseIP(fields[0]), net.ParseIP(fields[1]), parseResult(fields[2:]))
		if err != nil {
			return fmt.Errorf("build index fail: %s", err.Error())
		}
	}
	index.index = builder.Build()
	return nil
}
