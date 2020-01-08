package common

import "github.com/ontio/ontology/common"

type FileStore struct {
	FileHash       string
	FileDesc       string
	FileBlockCount uint64
	RealFileSize   uint64
	CopyNumber     uint64
	FirstPdp       bool
	PdpInterval    uint64
	TimeExpired    uint64
	PdpParam       []byte
	StorageType    uint64
}

type FileRenew struct {
	FileHash  string
	RenewTime uint64
}

type FileTransfer struct {
	FileHash string
	NewOwner common.Address
}
