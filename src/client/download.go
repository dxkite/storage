package client

import (
	"bytes"
	"dxkite.cn/go-storage/src/bitset"
	"dxkite.cn/go-storage/src/block"
	"dxkite.cn/go-storage/src/common"
	"dxkite.cn/go-storage/src/image"
	"dxkite.cn/go-storage/src/meta"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type Downloader struct {
	DownloadMeta
	File     *block.BlockFile
	BlockTry int
	Error    error
	mutex    sync.Mutex
}

type MetaDownloader struct {
	Downloader
	MetaPath string
	Check    bool
	Thread   int
}

func NewMetaDownloader(path string, check bool, num int) *MetaDownloader {
	m := &MetaDownloader{
		MetaPath: path,
		Check:    check,
		Thread:   num,
	}
	m.BlockTry = 5
	return m
}

func (d *MetaDownloader) DownloadToFile(path string) error {
	df, err := d.init(d.MetaPath, path)
	if err != nil {
		return err
	}
	if d.Downloaded == d.DownloadTotal {
		return nil
	}
	defer func() { _ = d.File.Close() }()
	if er := d.download(df, d.Thread); er != nil {
		return er
	}
	if d.Check && d.File.CheckSum() == false {
		return errors.New("hash check error")
	}
	_ = os.Remove(df)
	return err
}

func (d *Downloader) download(df string, max int) error {
	var g sync.WaitGroup
	var limit = make(chan bool, max)
	for _, bb := range d.Block {
		g.Add(1)
		go func(b meta.DataBlock) {
			limit <- true
			err := d.downloadBlock(df, &b)
			if err != nil {
				d.mutex.Lock()
				d.Error = err
				d.mutex.Unlock()
			}
			g.Done()
			<-limit
		}(bb)
	}
	g.Wait()
	return d.Error
}

func (d *MetaDownloader) initMeta(metaFile string) error {
	if strings.Index(metaFile, common.BASE_PROTOCOL+"://") == 0 {
		// storage://meta?dl=base64_encode(meta-url)
		m, er := meta.DecodeFromMetaProtocol(metaFile)
		if er != nil {
			return er
		}
		d.Info = *m
	} else if strings.Index(metaFile, "https://") == 0 || strings.Index(metaFile, "http://") == 0 {
		m, er := meta.DecodeFromUrl(metaFile)
		if er != nil {
			return er
		}
		d.Info = *m
	} else {
		m, er := meta.DecodeFromFile(metaFile)
		if er != nil {
			return er
		}
		d.Info = *m
	}
	if d.Status != meta.Finish {
		return errors.New("meta status error")
	}
	return nil
}

func (d *MetaDownloader) init(metaFile, p string) (string, error) {
	if er := d.initMeta(metaFile); er != nil {
		return "", er
	}

	log.Println("download meta", metaFile)
	log.Println(":name", d.Name)
	log.Println(":sha1", hex.EncodeToString(d.Hash))
	log.Println(":size", d.Size)
	log.Println("check enable", d.Check)

	_ = os.MkdirAll(p, os.ModePerm)
	df := path.Join(p, d.Name+common.EXT_DOWNLOADING)
	pp := path.Join(p, d.Name)

	if common.FileExist(df) {
		log.Println("reload download status")
		dd, err := DecodeToFile(df)
		if err != nil {
			return df, errors.New(fmt.Sprintf("reload downloading: %v", err))
		}
		d.DownloadMeta = *dd
		d.Downloaded = 0
		if er := d.parepareDownloadFile(pp); er != nil {
			return df, er
		}
	} else {
		d.Index = bitset.New(int64(len(d.Block)))
		d.DownloadTotal = len(d.Block)
		d.Downloaded = 0
		if er := d.parepareDownloadFile(pp); er != nil {
			return df, er
		}
	}
	return df, nil
}

func (d *MetaDownloader) checkBlock(bb meta.DataBlock) bool {
	start, end := d.calculateRange(bb.Index)
	rb := block.NewBlock(start, end, nil)
	if buf, err := d.File.ReadBlock(rb); err != nil {
		return false
	} else if bytes.Equal(ByteHash(buf), bb.Hash) {
		return true
	}
	return false
}

func (d *MetaDownloader) parepareDownloadFile(path string) error {
	flag := os.O_CREATE | os.O_RDWR
	exist := common.FileExist(path)

	if exist {
		log.Println("file exists", path)
	} else {
		flag |= os.O_TRUNC
		log.Println("create file", path)
	}

	file, err := os.OpenFile(path, flag, os.ModePerm)
	if err != nil {
		return err
	}

	d.File = &block.BlockFile{
		File: file,
		Hash: d.Hash,
	}

	if exist {
		log.Println("start check download blocks", path)
		for _, bb := range d.Block {
			if d.checkBlock(bb) {
				log.Println(fmt.Sprintf("block %d is downloaded", bb.Index))
				d.Index.Set(bb.Index)
				d.Downloaded++
			}
		}
	}
	return nil
}

type DownloadRetryable struct {
	try int
	d   *Downloader
}

func (r *DownloadRetryable) downloadBlock(dataBlock *meta.DataBlock) (block.Block, error) {
	// 下载成功
	var retry = false
	var err error
	buf, er := GetRemoteData(string(dataBlock.Data))
	if er != nil {
		retry = true
		err = er
	}

	bh := ByteHash(buf)
	if bytes.Equal(bh, dataBlock.Hash) == false {
		err = errors.New("hash check error")
		retry = true
	}

	if retry {
		r.try--
		// 可重试
		if r.try > 0 {
			log.Printf("block %d download error: %v, retry", dataBlock.Index, err)
			return r.downloadBlock(dataBlock)
		}
		return nil, err
	}

	log.Printf("block %d dwonloaded", dataBlock.Index)
	start, end := r.d.calculateRange(dataBlock.Index)
	return block.NewBlock(start, end, buf), err
}

func (d *Downloader) calculateRange(index int64) (begin, end int64) {
	begin = index * d.BlockSize
	end = begin + d.BlockSize
	if end > d.Size {
		end = d.Size
	}
	return begin, end
}

func (d *Downloader) IsDownloaded(dataBlock *meta.DataBlock) bool {
	if d.Index.Get(dataBlock.Index) {
		return true
	}
	return false
}

func (d *Downloader) downloadBlock(df string, dataBlock *meta.DataBlock) error {
	if d.IsDownloaded(dataBlock) {
		log.Printf("block %d is exist", dataBlock.Index)
		return nil
	}
	log.Printf("[%.2f%%] block %d downloading", float64(d.Downloaded)*100/float64(d.DownloadTotal), dataBlock.Index)
	dr := DownloadRetryable{d.BlockTry, d}
	bb, er := dr.downloadBlock(dataBlock)
	if er != nil {
		return er
	}
	if bb != nil {
		if wer := d.File.WriteBlock(bb); wer != nil {
			return wer
		}
		// 标记下载
		d.Downloaded++
		d.Index.Set(dataBlock.Index)
		log.Printf("[%.2f%%] block %d downloaded", float64(d.Downloaded)*100/float64(d.DownloadTotal), dataBlock.Index)
		if wer := EncodeToFile(df, &d.DownloadMeta); wer != nil {
			return wer
		}
	}
	return nil
}

func GetRemoteData(url string) ([]byte, error) {
	buf := &bytes.Buffer{}
	r, er := http.Get(url)
	if er != nil {
		return nil, er
	}
	defer func() { _ = r.Body.Close() }()
	if err := image.Decode(r.Body, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
