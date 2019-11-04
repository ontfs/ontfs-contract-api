package main

import (
	"encoding/hex"
	"flag"
	"github.com/ontio/ontfs-contract-api/client"
	"github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"os"
	"sync"
	"time"
)

const TestFileHash = "FileTest"

var fsClient *client.OntFsClient
var globalParam *ontfs.FsGlobalParam
var once sync.Once

var action = struct {
	getGlobalParam  bool
	getNodeInfoList bool
	getFileList     bool
	storeFile       bool
	getFileInfo     bool
	renewFile       bool
	delFile         bool
	getPdpInfoList  bool
	changeOwner     bool
	fileHash        string
	newOwner        string
}{}

func main() {
	flag.BoolVar(&action.getGlobalParam, "getGlobalParam", false, "getGlobalParam")
	flag.BoolVar(&action.getNodeInfoList, "getNodeInfoList", false, "getNodeInfoList")
	flag.BoolVar(&action.getFileList, "getFileList", false, "getFileList")
	flag.BoolVar(&action.storeFile, "storeFile", false, "storeFile")
	flag.BoolVar(&action.getFileInfo, "getFileInfo", false, "getFileInfo")
	flag.BoolVar(&action.renewFile, "renewFile", false, "renewFile")
	flag.BoolVar(&action.getPdpInfoList, "getPdpInfoList", false, "getPdpInfoList")
	flag.BoolVar(&action.delFile, "delFile", false, "delFile")
	flag.BoolVar(&action.changeOwner, "changeOwner", false, "changeOwner")
	flag.StringVar(&action.fileHash, "fileHash", TestFileHash, "   -fileHash")
	flag.StringVar(&action.newOwner, "newOwner", "", "   changeOwner - newOwner")
	flag.Parse()

	fsClient = client.Init("./wallet.dat", "pwd", "http://localhost:20336")
	if fsClient == nil {
		log.Error("Init error")
		return
	}

	if action.getGlobalParam {
		GetGlobalParam()
	} else if action.getNodeInfoList {
		GetNodeInfoList()
	} else if action.getFileList {
		GetFileList()
	} else if action.storeFile {
		StoreFile()
	} else if action.getFileInfo {
		GetFileInfo(action.fileHash)
	} else if action.renewFile {
		RenewFile(action.fileHash)
	} else if action.delFile {
		DeleteFile(action.fileHash)
	} else if action.changeOwner {
		ChangeOwner(action.fileHash, action.newOwner)
	} else if action.getPdpInfoList {
		GetPdpInfoList(action.fileHash)
	}
}

func GetGlobalParam() {
	var err error
	globalParam, err = fsClient.GetGlobalParam()
	if err != nil {
		log.Errorf("APP GetGlobalParam error: %s", err.Error())
		return
	}
	common.PrintStruct(*globalParam)
}

func GetNodeInfoList() {
	nodeInfoList, err := fsClient.GetNodeInfoList()
	if err != nil {
		log.Errorf("APP GetNodeInfoList error: %s", err.Error())
		return
	} else {
		log.Infof("NodeInfoListLen: %d", len(nodeInfoList.NodesInfo))
		for _, nodeInfo := range nodeInfoList.NodesInfo {
			common.PrintStruct(nodeInfo)
		}
	}
}

func GetFileList() {
	fileHashList, err := fsClient.GetFileList()
	if err != nil {
		log.Errorf("APP GetFileList error: %s", err.Error())
		return
	} else {
		for _, fileHash := range fileHashList.FilesH {
			log.Infof("FileHash: %s", string(fileHash.FHash))
		}
	}
}

func StoreFile() {
	txHash, err := fsClient.StoreFile(TestFileHash, 256, 1580000000, 1, []byte(TestFileHash),
		[]byte(TestFileHash), ontfs.FileStorageTypeUseFile, 256*256+256)
	if err != nil {
		log.Error("StoreFile error: ", err.Error())
		return
	}

	fsClient.PollForTxConfirmed(15*time.Second, txHash)

	fileInfo, err := fsClient.GetFileInfo(TestFileHash)
	if err != nil {
		log.Errorf("StoreFile GetFileInfo fileHash: %s error: %s", TestFileHash, err.Error())
		return
	} else if fileInfo == nil {
		log.Error("StoreFile GetFileInfo failed, fileInfo is nil")
		return
	}

	once.Do(func() {
		if err := connectFs(); err != nil {
			log.Error("connectFs error: ", err.Error())
			os.Exit(0)
		}
		log.Info("connection success")
	},
	)

	if err = sendToFs("StoreFile" + "|" + hex.EncodeToString(txHash)); err != nil {
		log.Error("sendToFs error: ", err.Error())
		return
	}
	closeConn()
}

func GetFileInfo(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		log.Error("GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	}
	common.PrintStruct(*fileInfo)

	filePdpNeedCount := (fileInfo.TimeExpired-fileInfo.TimeStart)/ontfs.DefaultPdpInterval + 1
	log.Infof("TotalPdpNeedCount: %d", filePdpNeedCount)
}

func RenewFile(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		log.Errorf("RenewFile GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	}
	renewTx, err := fsClient.RenewFile(fileHash, fileInfo.TimeExpired+1024)
	if err != nil {
		log.Error("RenewFile error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(14*time.Second, renewTx)

	fileInfo1, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		log.Errorf("RenewFile GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	}
	if fileInfo1.TimeExpired == fileInfo.TimeExpired+1024 {
		log.Info("RenewFile success")
	} else {
		log.Info("RenewFile failed")
	}
}

func DeleteFile(fileHash string) {
	delFileTx, err := fsClient.DeleteFiles([]string{fileHash})
	if err != nil {
		log.Error("DeleteFile error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(14*time.Second, delFileTx)

	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err == nil {
		log.Error("DeleteFile failed err is nil")
	} else if fileInfo != nil {
		log.Error("DeleteFile failed")
	} else {
		log.Info("DeleteFile success")
	}
}

func ChangeOwner(fileHash string, newOwner string) {
	newOwnerAddr, err := utils.AddressFromBase58(newOwner)
	if err != nil {
		log.Error("ChangeOwner AddressFromBase58 error: ", err.Error())
		return
	}

	ownerChangeTx, err := fsClient.ChangeFileOwner(fileHash, newOwnerAddr)
	if err != nil {
		log.Error("ChangeOwner error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(14*time.Second, ownerChangeTx)

	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err == nil {
		log.Error("ChangeOwner failed err is nil")
		return
	}
	if fileInfo.FileOwner == newOwnerAddr {
		log.Infof("ChangeOwner success")
	} else {
		log.Infof("ChangeOwner failed")
	}
}

func GetPdpInfoList(fileHash string) {
	pdpRecordList, err := fsClient.GetFilePdpRecordList(fileHash)
	if err != nil {
		log.Errorf("APP GetFilePdpRecordList error: %s", err.Error())
		return
	} else {
		for _, pdpRecord := range pdpRecordList.PdpRecords {
			common.PrintStruct(pdpRecord)
		}
	}
}
