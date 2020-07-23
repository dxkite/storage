package downloader

import (
	"bytes"
	"dxkite.cn/storage/bitset"
	"dxkite.cn/storage/block"
	"dxkite.cn/storage/common"
	"dxkite.cn/storage/image"
	"dxkite.cn/storage/meta"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

type Downloader struct {
	ProcessMeta
	File     *block.BlockFile
	BlockTry int
	Error    error
	mutex    sync.Mutex
	// 进度
	Processor Processor
}

func (d *Downloader) download(max int) error {
	var g sync.WaitGroup
	var limit = make(chan bool, max)
	for _, bb := range d.Block {
		g.Add(1)
		go func(b meta.DataBlock) {
			limit <- true
			err := d.downloadBlock(&b)
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

// 初始化
func (d *Downloader) init(p Processor, stream block.File) error {
	d.Processor = p
	dd, _ := d.Processor.Load()
	if dd == nil {
		d.Index = bitset.New(int64(len(d.Block)))
		d.DownloadTotal = len(d.Block)
		d.Downloaded = 0
	} else {
		d.ProcessMeta = *dd
	}

	d.File = &block.BlockFile{
		File: stream,
		Hash: d.Hash,
	}

	log.Println("start check download blocks")
	for _, bb := range d.Block {
		if d.checkBlock(bb) {
			log.Println(fmt.Sprintf("block %d is downloaded", bb.Index))
			d.Index.Set(bb.Index)
			d.Downloaded++
		}
	}
	return nil
}

// 检测快是否可用
func (d *Downloader) checkBlock(bb meta.DataBlock) bool {
	start, end := d.calculateRange(bb.Index)
	rb := block.NewBlock(start, end, nil)
	if buf, err := d.File.ReadBlock(rb); err != nil {
		return false
	} else if bytes.Equal(common.ByteHash(buf), bb.Hash) {
		return true
	}
	return false
}

// 自动重试下载
type DownloadRetryable struct {
	try int
	d   *Downloader
}

// 下载数据
func (r *DownloadRetryable) downloadBlock(dataBlock *meta.DataBlock) (block.Block, error) {
	// 下载成功
	var retry = false
	var err error

	buf, er := r.d.GetBlockData(string(dataBlock.Data), meta.EncodeType(r.d.Encode), meta.DataType(r.d.Type))
	if er != nil {
		retry = true
		err = er
	}

	if err == nil {
		bh := common.ByteHash(buf)
		if bytes.Equal(bh, dataBlock.Hash) == false {
			err = errors.New("hash check error")
			retry = true
		}
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

// 计算块位置
func (d *Downloader) calculateRange(index int64) (begin, end int64) {
	begin = index * d.BlockSize
	end = begin + d.BlockSize
	if end > d.Size {
		end = d.Size
	}
	return begin, end
}

// 检测块是否下载成功
func (d *Downloader) IsDownloaded(dataBlock *meta.DataBlock) bool {
	if d.Index.Get(dataBlock.Index) {
		return true
	}
	return false
}

// 下载块数据
func (d *Downloader) downloadBlock(dataBlock *meta.DataBlock) error {
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
		if wer := d.Processor.Save(&d.ProcessMeta, dataBlock.Index, bb.Start(), bb.End()); wer != nil {
			return wer
		}
	}
	return nil
}

// 获取块数据
func (d *Downloader) GetBlockData(url string, encode meta.EncodeType, dt meta.DataType) ([]byte, error) {
	var r io.Reader
	buf := &bytes.Buffer{}

	if dt == meta.Type_Stream {
		r = strings.NewReader(url)
	} else {
		rr, er := common.HttpGet(url)
		if er != nil {
			return nil, er
		}
		defer func() { _ = rr.Close() }()
		r = rr
	}
	if encode == meta.Encode_Image {
		if err := image.Decode(r, buf); err != nil {
			return nil, err
		}
	} else {
		if _, err := io.Copy(buf, r); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (d *Downloader) MetaFromFile(path string) error {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	return d.LoadFromStream(f)
}

func (d *Downloader) MetaFromProtocol(path string) error {
	u, er := url.Parse(path)
	if er != nil {
		return er
	}
	if u.Scheme != common.BASE_PROTOCOL {
		return errors.New("need protocol " + common.BASE_PROTOCOL)
	}
	if u.Host != common.HOST_META {
		return errors.New("need host meta")
	}
	dl := u.Query().Get("dl")
	if dd, err := base64.StdEncoding.DecodeString(dl); err != nil {
		return errors.New(fmt.Sprintf("protocol decode dl %v", err))
	} else {
		return d.LoadFromUrl(string(dd))
	}
}

func (d *Downloader) LoadFromUrl(url string) error {
	if res, er := common.HttpGet(url); er != nil {
		return er
	} else {
		defer func() { _ = res.Close() }()
		if bb, er := ioutil.ReadAll(res); er != nil {
			return er
		} else {
			return d.LoadFromStream(bytes.NewReader(bb))
		}
	}
}

// 从流载入
func (d *Downloader) LoadFromStream(s io.Reader) error {
	m, er := meta.DecodeFromStream(s)
	if er != nil {
		return er
	}
	d.Info = *m
	if d.Status != meta.Finish {
		return errors.New("meta status error")
	}
	return nil
}

func (d *Downloader) Load(metaFile string) error {
	if strings.Index(metaFile, common.BASE_PROTOCOL+"://") == 0 {
		// storage://meta?dl=base64_encode(meta-url)
		return d.MetaFromProtocol(metaFile)
	} else if strings.Index(metaFile, "https://") == 0 || strings.Index(metaFile, "http://") == 0 {
		return d.LoadFromUrl(metaFile)
	} else {
		return d.MetaFromFile(metaFile)
	}
}

type MetaDownloader struct {
	Downloader
	Check  bool
	Thread int
}

func NewMetaDownloader(check bool, num int) *MetaDownloader {
	m := &MetaDownloader{
		Check:  check,
		Thread: num,
	}
	m.BlockTry = 5
	return m
}

// 下载到文件
func (d *MetaDownloader) DownloadToFile(p string) error {
	log.Println(":name", d.Name)
	log.Println(":sha1", hex.EncodeToString(d.Hash))
	log.Println(":size", d.Size)
	log.Println("check enable", d.Check)
	_ = os.MkdirAll(p, os.ModePerm)
	df := path.Join(p, d.Name+common.EXT_DOWNLOADING)
	pp := path.Join(p, d.Name)
	if f, er := d.getOutputFile(pp); er != nil {
		return er
	} else {
		return d.DownloadToStream(LocalProcessor{df}, f)
	}
}

// 下载到流
func (d *MetaDownloader) DownloadToStream(p Processor, file block.File) error {
	if err := d.init(p, file); err != nil {
		return err
	}
	if d.Downloaded == d.DownloadTotal {
		return nil
	}
	defer func() { _ = d.File.Close() }()
	if err := d.download(d.Thread); err != nil {
		return err
	}
	if d.Check && d.File.CheckSum() == false {
		err := errors.New("hash check error")
		return err
	}
	_ = d.Processor.Clear()
	return nil
}

// 获取下载的文件
func (d *MetaDownloader) getOutputFile(path string) (block.File, error) {
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
		return nil, err
	}
	return file, nil
}
