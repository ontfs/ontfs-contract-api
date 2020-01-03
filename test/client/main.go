package main

import (
	"encoding/hex"
	"flag"
	"os"
	"sync"
	"time"
	"fmt"

	"github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontfs-contract-api/core"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"

)

const TestFileHash = "FileTest"
const DefaultPdpInterval = 4 * 60 * 60

var fsClient *core.Core
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
	readFile        bool
	getPdpInfoList  bool
	changeOwner     bool
	createSpace     bool
	updateSpace     bool
	deleteSpace     bool
	getSpaceInfo    bool
	fileHash        string
	newOwner        string
}{}

func main() {
	log.InitLog(log.InfoLog, ".")
	flag.BoolVar(&action.getGlobalParam, "getGlobalParam", false, "getGlobalParam")
	flag.BoolVar(&action.getNodeInfoList, "getNodeInfoList", false, "getNodeInfoList")
	flag.BoolVar(&action.getPdpInfoList, "getPdpInfoList", false, "getPdpInfoList")

	flag.BoolVar(&action.getFileList, "getFileList", false, "getFileList")

	flag.BoolVar(&action.storeFile, "storeFile", false, "storeFile")
	flag.BoolVar(&action.getFileInfo, "getFileInfo", false, "getFileInfo")
	flag.BoolVar(&action.renewFile, "renewFile", false, "renewFile")
	flag.BoolVar(&action.delFile, "delFile", false, "delFile")
	flag.BoolVar(&action.readFile, "readFile", false, "readFile")
	flag.BoolVar(&action.changeOwner, "changeOwner", false, "changeOwner")

	flag.BoolVar(&action.createSpace, "createSpace", false, "createSpace")
	flag.BoolVar(&action.updateSpace, "updateSpace", false, "updateSpace")
	flag.BoolVar(&action.deleteSpace, "deleteSpace", false, "deleteSpace")
	flag.BoolVar(&action.getSpaceInfo, "getSpaceInfo", false, "getSpaceInfo")

	flag.StringVar(&action.fileHash, "fileHash", TestFileHash, "   -fileHash")
	flag.StringVar(&action.newOwner, "newOwner", "", "   changeOwner - newOwner")
	flag.Parse()

	fsClient = core.Init("./wallet.dat", "pwd", "http://127.0.0.1:20336", 0, 20000)
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
	} else if action.readFile {
		ReadFile(action.fileHash)
	} else if action.changeOwner {
		TransferFile(action.fileHash, action.newOwner)
	} else if action.getPdpInfoList {
		GetPdpInfoList(action.fileHash)
	} else if action.createSpace {
		CreateSpace()
	} else if action.getSpaceInfo {
		GetSpaceInfo()
	} else if action.deleteSpace {
		DeleteSpace()
	} else if action.updateSpace {
		UpdateSpace()
	}
}

func CreateSpace() {
	timeExpired := uint64(time.Now().Unix()) + 3600*24
	txHash, err := fsClient.CreateSpace(1024*1024, 3, DefaultPdpInterval, timeExpired)
	if err != nil {
		log.Error("CreateSpace error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(15*time.Second, txHash)
	GetSpaceInfo()
}

func GetSpaceInfo() {
	spaceInfo, err := fsClient.GetSpaceInfo()
	if err != nil {
		log.Error("GetSpaceInfo error: ", err.Error())
		return
	}
	common.PrintStruct(*spaceInfo)
}

func DeleteSpace() {
	txHash, err := fsClient.DeleteSpace()
	if err != nil {
		log.Error("DeleteSpace error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(15*time.Second, txHash)
	spaceInfo, err := fsClient.GetSpaceInfo()
	if err != nil && spaceInfo == nil {
		log.Info("DeleteSpace success")
	} else {
		log.Error("DeleteSpace failed")
	}
}

func UpdateSpace() {
	spaceInfo1, err := fsClient.GetSpaceInfo()
	if err != nil {
		log.Error("UpdateSpace GetSpaceInfo1 error: ", err.Error())
		return
	}
	common.PrintStruct(*spaceInfo1)

	timeExpired := uint64(time.Now().Unix()) + 3600*24
	txHash, err := fsClient.UpdateSpace(1024*2048, timeExpired)
	if err != nil {
		log.Error("UpdateSpace error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(15*time.Second, txHash)

	spaceInfo2, err := fsClient.GetSpaceInfo()
	if err != nil {
		log.Error("UpdateSpace GetSpaceInfo2 error: ", err.Error())
		return
	}
	common.PrintStruct(*spaceInfo2)

	if spaceInfo1.Volume != spaceInfo2.Volume || spaceInfo1.TimeExpired != spaceInfo2.TimeExpired {
		log.Info("UpdateSpace Success")
	} else {
		log.Error("UpdateSpace Failed")
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
	nodeInfoList, err := fsClient.GetNodeInfoList(9999999)
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
	timeExpired := uint64(time.Now().Unix()) + 3600
	fileStores := []common.FileStore{
		{
			FileHash:      TestFileHash,
			FileDesc:       TestFileHash,
			FileBlockCount: 256,
			RealFileSize:   256*256 + 256,
			CopyNumber:     3,
			PdpInterval:    DefaultPdpInterval,
			TimeExpired:    timeExpired,
			PdpParam:       []byte(TestFileHash),
			StorageType:    ontfs.FileStorageTypeUseFile,
		},
	}

	_, err, storeErrors := fsClient.StoreFiles(fileStores)
	if err != nil {
		log.Error("StoreFile error: ", err.Error())
		return
	}

	if len(storeErrors.ObjectErrors) == 0 {
		fmt.Printf("StoreFile no object error\n")
		return
	}
	for k, v := range storeErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}

	once.Do(func() {
		if err := connectFs(); err != nil {
			log.Error("connectFs error: ", err.Error())
			os.Exit(0)
		}
		log.Info("connection success")
	},
	)

	if err = sendToFs("StoreFile" + "|" + TestFileHash); err != nil {
		log.Error("sendToFs error: ", err.Error())
		return
	}
	closeConn()
}

func GetFileInfo(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		log.Errorf("GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	}
	common.PrintStruct(*fileInfo)

	filePdpNeedCount := (fileInfo.TimeExpired-fileInfo.TimeStart)/fileInfo.PdpInterval + 1
	log.Infof("TotalPdpNeedCount: %d", filePdpNeedCount)
}

func RenewFile(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		log.Errorf("RenewFile GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	}

	fileRenew := []common.FileRenew{
		{
			FileHash:      fileHash,
			RenewTime:     fileInfo.TimeExpired+1024,
		},
	}
	_, err, renewErrors := fsClient.RenewFiles(fileRenew)
	if err != nil {
		log.Error("RenewFile error: ", err.Error())
		return
	}

	if len(renewErrors.ObjectErrors) == 0 {
		fmt.Printf("RenewFiles no object error\n")
		return
	}
	for k, v := range renewErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}
}

func DeleteFile(fileHash string) {
	_, err, delErrors := fsClient.DeleteFiles([]string{fileHash})
	if err != nil {
		log.Error("DeleteFile error: ", err.Error())
		return
	}

	if len(delErrors.ObjectErrors) == 0 {
		fmt.Printf("DeleteFile no object error\n")
		return
	}
	for k, v := range delErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}

}

func TransferFile(fileHash string, newOwner string) {
	newOwnerAddr, err := utils.AddressFromBase58(newOwner)
	if err != nil {
		log.Error("ChangeOwner AddressFromBase58 error: ", err.Error())
		return
	}

	fileTransfer := []common.FileTransfer{
		{
			FileHash:      fileHash,
			NewOwner:      newOwnerAddr,
		},
	}

	_, err, transferErrors := fsClient.TransferFiles(fileTransfer)
	if err != nil {
		log.Error("ChangeOwner error: ", err.Error())
		return
	}
	if len(transferErrors.ObjectErrors) == 0 {
		fmt.Printf("TransferFile no object error\n")
		return
	}
	for k, v := range transferErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
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

func ReadFile(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		log.Errorf("StoreFile GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	} else if fileInfo == nil {
		log.Error("StoreFile GetFileInfo failed, fileInfo is nil")
		return
	}

	pdpRecordList, err := fsClient.GetFilePdpRecordList(fileHash)
	if err != nil {
		log.Errorf("APP GetFilePdpRecordList error: %s", err.Error())
		return
	} else {
		for _, pdpRecord := range pdpRecordList.PdpRecords {
			common.PrintStruct(pdpRecord)
		}
	}

	readPlans := []ontfs.ReadPlan{
		{
			NodeAddr:         pdpRecordList.PdpRecords[0].NodeAddr,
			MaxReadBlockNum:  fileInfo.FileBlockCount,
			HaveReadBlockNum: 0,
		},
	}
	readTx, err := fsClient.FileReadPledge(fileHash, readPlans)
	if err != nil {
		log.Error("FileReadPledge failed error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(14*time.Second, readTx)

	haveReadBlockNum := uint64(0)
	readPledge, err := fsClient.GetFileReadPledge(fileHash, fsClient.WalletAddr)
	if err != nil {
		log.Error("GetFileReadPledge failed error: ", err.Error())
		return
	}
	for _, readPlan := range readPledge.ReadPlans {
		if readPlan.NodeAddr == fsClient.WalletAddr {
			haveReadBlockNum = readPlan.HaveReadBlockNum
		}
	}

	once.Do(func() {
		if err := connectFs(); err != nil {
			log.Error("connectFs error: ", err.Error())
			os.Exit(0)
		}
		log.Info("connection success")
	},
	)

	if err = sendToFs("ReadFile" + "|" + fileHash + "|" + fsClient.WalletAddr.ToBase58()); err != nil {
		log.Error("sendToFs error: ", err.Error())
		return
	}

	for i := uint64(0); i < fileInfo.FileBlockCount; i++ {
		fileReadSlice, err := fsClient.GenFileReadSettleSlice([]byte(fileHash), readPlans[0].NodeAddr, i+haveReadBlockNum, 1)
		if err != nil {
			log.Errorf("GenFileReadSettleSlice error: %s", err.Error())
			return
		}
		sliceData := common.FileReadSettleSliceSerialize(fileReadSlice)
		sliceString := hex.EncodeToString(sliceData)

		log.Info("sendToFs FileReadSettleSlice")
		sendToFs(sliceString)
	}
	closeConn()
}
