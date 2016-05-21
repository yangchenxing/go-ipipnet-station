# go-ipipnet-station

[![Build Status](https://travis-ci.org/yangchenxing/go-ipipnet-station.svg?branch=master)](https://travis-ci.org/yangchenxing/go-ipipnet-station)
[![GoDoc](http://godoc.org/github.com/yangchenxing/go-ipipnet-station?status.svg)](http://godoc.org/github.com/yangchenxing/go-ipipnet-station)

IPIPNet基站IP判别库

## Example

    index := &Index{
        Downloader: downloader.Downloader{
            LocalPath: "StationGWList.sample.txt",
        },
    }
    index.Initialize()
    result, _ := index.Search(net.ParseIP("001.050.014.001"))
    
关于下载器和索引机制，请查看[go-ipipnet-downloader](https://github.com/yangchenxing/go-ipipnet-downloader)和[go-ip-index](https://github.com/yangchenxing/go-ip-index)
