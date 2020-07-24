package storage

import (
	"dxkite.cn/storage/bitset"
	"dxkite.cn/storage/meta"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

type ProcessStatus int

const (
	PROCESS_START ProcessStatus = iota
	PROCESS_DONE
	PROCESS_EXIST
	PROCESS_ERROR
)

func (s ProcessStatus) String() string {
	switch s {
	case PROCESS_START:
		return "PROCESS_START"
	case PROCESS_DONE:
		return "PROCESS_DONE"
	case PROCESS_ERROR:
		return "PROCESS_ERROR"
	case PROCESS_EXIST:
		return "PROCESS_EXIST"
	}
	return "ProcessStatus:<" + strconv.Itoa(int(s)) + ">"
}

type DownloadInfo struct {
	Index         bitset.BitSet
	Downloaded    int
	DownloadTotal int
}

func (info *DownloadInfo) EncodeToFile(path string) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewEncoder(f)
	return b.Encode(info)
}

func (info *DownloadInfo) DecodeFromFile(path string) error {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewDecoder(f)
	der := b.Decode(&info)
	if der != nil {
		return der
	}
	return nil
}

type FileProcessor interface {
	EncodeToFile(path string) error
	DecodeFromFile(path string) error
}

// 下载进度
type DownloadProcessor interface {
	// 加载下载进度
	Load() (*DownloadInfo, error)
	// 保存下载进度
	Save(meta *DownloadInfo) error
	// 通知下载进度
	Process(status ProcessStatus, index, start, end int64, err error) error
	// 下载成功
	Finish() error
}

// 文件下载进度保存
type FileDownloadProcessor struct {
	p string
}

func NewDownloadProcessor(p string) *FileDownloadProcessor {
	return &FileDownloadProcessor{
		p: p,
	}
}

func (p *FileDownloadProcessor) Load() (*DownloadInfo, error) {
	if !FileExist(p.p) {
		return nil, errors.New("download meta not exists")
	}
	dd := new(DownloadInfo)
	err := dd.DecodeFromFile(p.p)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("reload downloading: %v", err))
	}
	return dd, nil
}

func (p *FileDownloadProcessor) Finish() error {
	return os.Remove(p.p)
}

func (p *FileDownloadProcessor) Save(info *DownloadInfo) error {
	if wer := info.EncodeToFile(p.p); wer != nil {
		return wer
	}
	return nil
}

func (p *FileDownloadProcessor) Process(status ProcessStatus, index, start, end int64, err error) error {
	log.Println("download", status, index, start, end, err)
	return nil
}

type UploadInfo struct {
	Index bitset.BitSet
	Meta  *meta.Info
}

func NewUploadInfo(name string, size, block, blockSize int64, info []byte) *UploadInfo {
	m := &meta.Info{
		Hash:      info,
		BlockSize: blockSize,
		Size:      size,
		Name:      name,
		Status:    meta.Create,
		Type:      int32(meta.Type_URI),
		Encode:    int32(meta.Encode_Image),
		Block:     []meta.DataBlock{},
	}
	return &UploadInfo{
		Index: bitset.New(block),
		Meta:  m,
	}
}

func (info *UploadInfo) EncodeToFile(path string) error {
	f, er := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewEncoder(f)
	return b.Encode(info)
}

func (info *UploadInfo) DecodeFromFile(path string) error {
	f, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if er != nil {
		return er
	}
	defer func() { _ = f.Close() }()
	b := gob.NewDecoder(f)
	der := b.Decode(&info)
	if der != nil {
		return der
	}
	return nil
}

// 上传进度
type UploadProcessor interface {
	// 加载上传进度
	Load() (*UploadInfo, error)
	// 保存上传进度
	Save(meta *UploadInfo) error
	// 通知上传进度
	Process(status ProcessStatus, index, start, end int64, err error) error
	// 上传成功
	Finish() error
}

// 文件下载进度保存
type FileUploadProcessor struct {
	p    string
	info *UploadInfo
}

func NewFileUploadProcessor(p string, info *UploadInfo) *FileUploadProcessor {
	return &FileUploadProcessor{
		p:    p,
		info: info,
	}
}

func (p *FileUploadProcessor) Load() (*UploadInfo, error) {
	if !FileExist(p.p) {
		return p.info, nil
	}
	dd := new(UploadInfo)
	err := dd.DecodeFromFile(p.p)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("reload downloading: %v", err))
	}
	return dd, nil
}

func (p *FileUploadProcessor) Save(info *UploadInfo) error {
	if wer := info.EncodeToFile(p.p); wer != nil {
		return wer
	}
	return nil
}

func (p *FileUploadProcessor) Process(status ProcessStatus, index, start, end int64, err error) error {
	log.Println("upload", status, index, start, end, err)
	return nil
}

func (p *FileUploadProcessor) Finish() error {
	return os.Remove(p.p)
}
