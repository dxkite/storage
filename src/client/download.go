package client

import (
	"bytes"
	"context"
	"dxkite.cn/go-storage/src/bitset"
	"dxkite.cn/go-storage/src/block"
	"dxkite.cn/go-storage/src/image"
	"dxkite.cn/go-storage/src/meta"
	"dxkite.cn/go-storage/storage"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type Downloader struct {
	DownloadMeta
	File  *block.BlockFile
	Error error
	mutex sync.Mutex
}

type MetaDownloader struct {
	Downloader
	MetaPath string
}

type RemoteDownloader struct {
	Downloader
	Remote string
}

func NewRemoteDownloader(remote string, info []byte) *RemoteDownloader {
	d := &RemoteDownloader{
		Remote: remote,
	}
	d.Hash = info
	return d
}

func NewMetaDownloader(path string) *MetaDownloader {
	return &MetaDownloader{
		MetaPath: path,
	}
}

func (d *RemoteDownloader) DownloadToFile(path string) error {
	df, err := d.init(path)
	if err != nil {
		return err
	}
	return d.download(df)
}

func (d *MetaDownloader) DownloadToFile(path string) error {
	df, err := d.init(d.MetaPath, path)
	if err != nil {
		return err
	}
	return d.download(df)
}

func (d *Downloader) download(df string) error {
	var g sync.WaitGroup
	for _, bb := range d.Block {
		g.Add(1)
		go func(b meta.DataBlock) {
			err := d.downloadBlock(df, &b)
			if err != nil {
				log.Println("error", err)
			}
			g.Done()
		}(bb)
	}
	g.Wait()
	return nil
}

func (d *MetaDownloader) init(metaFile, p string) (string, error) {
	m, er := meta.DecodeToFile(metaFile)
	if er != nil {
		return "", er
	}
	d.MetaInfo = *m
	_ = os.MkdirAll(p, os.ModePerm)
	df := path.Join(p, fmt.Sprintf("%x.gs-downloading", d.Hash))
	if FileExist(df) {
		log.Println("reload meta info")
		dd, err := DecodeToFile(df)
		if err != nil {
			return df, errors.New(fmt.Sprintf("reload downloading: %v", err))
		}
		d.DownloadMeta = *dd
		file, err := os.OpenFile(path.Join(p, d.Name), os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			return df, err
		}
		d.File = &block.BlockFile{
			File: file,
			Hash: d.Hash,
		}
	} else {
		d.Index = bitset.New(int64(len(m.Block)))
		d.DownloadTotal = len(m.Block)
		d.Downloaded = 0
		log.Println("create file", path.Join(p, d.Name))
		file, err := os.OpenFile(path.Join(p, d.Name), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return df, err
		}
		d.File = &block.BlockFile{
			File: file,
			Hash: m.Hash,
		}
	}
	return df, nil
}

func (d *RemoteDownloader) init(p string) (string, error) {
	_ = os.MkdirAll(p, os.ModePerm)
	df := path.Join(p, fmt.Sprintf("%x.gs-downloading", d.Hash))

	if FileExist(df) {
		log.Println("reload meta info")
		dd, err := DecodeToFile(df)
		if err != nil {
			return df, errors.New(fmt.Sprintf("reload downloading: %v", err))
		}
		d.DownloadMeta = *dd
		file, err := os.OpenFile(path.Join(p, d.Name), os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			return df, err
		}
		d.File = &block.BlockFile{
			File: file,
			Hash: d.Hash,
		}
	} else {
		log.Println("downloading meta info")
		m, er := d.getMeta()
		if er != nil {
			return df, er
		}

		d.MetaInfo = *NewMeta(m)
		d.Index = bitset.New(int64(len(m.Blocks)))
		d.DownloadTotal = len(m.Blocks)
		d.Downloaded = 0

		log.Println("create file", path.Join(p, d.Name))
		file, err := os.OpenFile(path.Join(p, d.Name), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return df, err
		}
		d.File = &block.BlockFile{
			File: file,
			Hash: m.Hash,
		}
	}
	return df, nil
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) { // 根据错误类型进行判断
			return true
		}
		return false
	}
	return true
}

type DownloadRetryable struct {
	try int
	d   *Downloader
}

func (r *DownloadRetryable) downloadBlock(dataBlock *meta.DataBlock) (block.Block, error) {
	// 下载成功
	if r.d.Index.Get(dataBlock.Index) {
		log.Printf("block %d is exist", dataBlock.Index)
		return nil, nil
	}
	var retry = false
	var err error
	buf, er := GetRemoteData(string(dataBlock.Data))
	if er != nil {
		retry = true
		err = er
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
	block.NewBlock(start, end, buf)
	return block.NewBlock(start, end, buf), err
}

func (d *Downloader) downloadBlock(df string, dataBlock *meta.DataBlock) error {
	log.Printf("[%.2f%%] block %d downloading", float64(d.Downloaded)*100/float64(d.DownloadTotal), dataBlock.Index)
	dr := DownloadRetryable{5, d}
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

func (d *RemoteDownloader) getMeta() (*storage.DataResponse, error) {
	conn, err := grpc.Dial(d.Remote, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()
	c := storage.NewGoStorageClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	r, er := c.Get(ctx, &storage.GetResponse{Info: d.Hash})
	if er != nil {
		return nil, er
	}
	return r, nil
}

func (d *Downloader) calculateRange(index int64) (begin, end int64) {
	begin = index * d.BlockSize
	end = begin + d.BlockSize
	if end > d.Size {
		end = d.Size
	}
	return begin, end
}

func (d *Downloader) createBlock(index int64, info []byte, url string) (block.Block, error) {
	b, err := GetRemoteData(url)
	if err != nil {
		return nil, err
	}
	start, end := d.calculateRange(index)
	return block.NewBlock(start, end, b), nil
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
