package server

type MetaInfo struct {
	Hash   []byte
	Size   int64
	Encode int32
	Blocks []DataBlock
}

type DataBlock struct {
	Hash  []byte
	Type  int32
	Index int64
	Data  []byte
}
