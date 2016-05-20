package stationip

import (
	"bufio"
	"container/list"
	"fmt"
	"github.com/yangchenxing/go-regionid"
	"io"
	"net"
	"strings"
)

type NetType int

const (
	UnknownNetType NetType = iota
	Station
	WiFi
)

var (
	netTypeName = map[NetType]string{
		UnknownNetType: "Unknown",
		Station:        "Station",
		WiFi:           "Wi-Fi",
	}
)

func (nt NetType) String() string {
	return netTypeName[nt]
}

var (
	minBinarySearchRange = 10
)

type section struct {
	Lower  uint32
	Upper  uint32
	Result Result
}

type index struct {
	Sections []section
	Index    [256][2]int
	ETag     string
}

type Result struct {
	NetType  NetType
	Location regionid.Location
	Owner    string
	Operator string
}

var (
	notFound = Result{
		NetType:  UnknownNetType,
		Location: nil,
		Owner:    "",
		Operator: "",
	}
)

func (idx *index) search(ip net.IP) Result {
	key := ip2Int(ip.To4())
	idxKey := int(key >> 24)
	left, right := idx.Index[idxKey][0], idx.Index[idxKey][1]
	if left == -1 && right == -1 {
		return notFound
	}
	for right-left > minBinarySearchRange {
		mid := (right + left) / 2
		sec := idx.Sections[mid]
		if sec.Lower > key {
			right = mid - 1
		} else if sec.Upper < key {
			left = mid + 1
		} else {
			return sec.Result
		}
	}
	for i := left; i <= right; i++ {
		sec := idx.Sections[i]
		if sec.Lower <= key && sec.Upper >= key {
			return sec.Result
		}
	}
	return notFound
}

func newIndex(r io.Reader, etag string) (*index, error) {
	br := bufio.NewScanner(r)
	sections := list.New()
	var idx [256][2]int
	for i := 0; i < 256; i++ {
		idx[i][0], idx[i][1] = -1, -1
	}
	var sec section
	lastIdxKey := -1
	for lineID := 1; br.Scan(); lineID++ {
		text := strings.TrimSpace(br.Text())
		if len(text) == 0 {
			continue
		}
		if newSec, err := parseSection(text); err != nil {
			return nil, fmt.Errorf("parse section fail at line %d: %s", lineID, err.Error())
		} else if sec.Upper+1 == newSec.Lower && (sec.Lower>>24) == (newSec.Lower>>24) && sec.Result == newSec.Result {
			// merge
			sec.Upper = newSec.Upper
			continue
		} else {
			if sec.Lower > 0 {
				idxKey := int(sec.Lower >> 24)
				idx[idxKey][1] = sections.Len()
				if idxKey != lastIdxKey {
					idx[idxKey][0] = sections.Len()
				}
				sections.PushBack(sec)
				lastIdxKey = idxKey
			}
			sec = newSec
		}
	}
	if sec.Lower > 0 {
		idxKey := int(sec.Lower >> 24)
		idx[idxKey][1] = sections.Len()
		if idxKey != lastIdxKey {
			idx[idxKey][0] = sections.Len()
		}
		sections.PushBack(sec)
		lastIdxKey = idxKey
	}
	newIndex := &index{
		Sections: make([]section, sections.Len()),
		Index:    idx,
		ETag:     etag,
	}
	for i, sec := 0, sections.Front(); sec != nil; i, sec = i+1, sec.Next() {
		newIndex.Sections[i] = sec.Value.(section)
	}
	return newIndex, nil
}

func parseSection(s string) (sec section, err error) {
	p := strings.Split(strings.TrimSpace(s), "\t")
	if len(p) != 8 {
		err = fmt.Errorf("bad formatted data line: %q", s)
		return
	}
	sec.Lower = parseIP2Int(p[0])
	sec.Upper = parseIP2Int(p[1])
	sec.Result.Location = regionid.GetLocation(p[2], p[3], p[4])
	if p[5] == "*" {
		sec.Result.Owner = p[6]
	} else {
		sec.Result.Owner = p[5]
	}
	sec.Result.Operator = p[6]
	switch p[7] {
	case "WIFI":
		sec.Result.NetType = WiFi
	case "基站":
		sec.Result.NetType = Station
	default:
		err = fmt.Errorf("unknown net type: %q", p[7])
		return
	}
	return
}

func parseIP2Int(s string) uint32 {
	ip := net.ParseIP(s).To4()
	return ip2Int(ip)
}

func ip2Int(ip net.IP) uint32 {
	return (uint32(ip[0]) << 24) | (uint32(ip[1]) << 16) | (uint32(ip[2]) << 8) | uint32(ip[3])
}
