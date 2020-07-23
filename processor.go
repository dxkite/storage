package storage

import (
	"errors"
	"fmt"
	"os"
)

// 下载进度处理
type Processor interface {
	// 加载下载进度
	Load() (*ProcessMeta, error)
	// 移除下载进度
	Clear() error
	// 保存下载进度
	Save(meta *ProcessMeta, index, start, end int64) error
}

type LocalProcessor struct {
	p string
}

func (p LocalProcessor) Load() (*ProcessMeta, error) {
	if !FileExist(p.p) {
		return nil, errors.New("download meta not exists")
	}
	dd, err := DecodeToFile(p.p)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("reload downloading: %v", err))
	}
	return dd, nil
}

func (p LocalProcessor) Clear() error {
	return os.Remove(p.p)
}

func (p LocalProcessor) Save(meta *ProcessMeta, index, start, end int64) error {
	if wer := EncodeToFile(p.p, meta); wer != nil {
		return wer
	}
	return nil
}
