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
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"

)

const TestFileHash = "FileTest"
const DefaultPdpInterval = 4 * 60 * 60

var fsClient *core.Core
var globalParam *ontfs.FsGlobalParam
var once sync.Once

var action = struct {
	rpcAddr         string
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
	flag.StringVar(&action.rpcAddr, "rpcAddr", "", "rpcAddr")
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

	fsClient = core.Init("./wallet.dat", "pwd", action.rpcAddr, 20000, 20000)
	if fsClient == nil {
		fmt.Println("Init error")
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
		fmt.Println("CreateSpace error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(15*time.Second, txHash)
	GetSpaceInfo()
}

func GetSpaceInfo() {
	spaceInfo, err := fsClient.GetSpaceInfo()
	if err != nil {
		fmt.Println("GetSpaceInfo error: ", err.Error())
		return
	}
	common.PrintStruct(*spaceInfo)
}

func DeleteSpace() {
	txHash, err := fsClient.DeleteSpace()
	if err != nil {
		fmt.Println("DeleteSpace error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(15*time.Second, txHash)
	spaceInfo, err := fsClient.GetSpaceInfo()
	if err != nil && spaceInfo == nil {
		fmt.Println("DeleteSpace success")
	} else {
		fmt.Println("DeleteSpace failed")
	}
}

func UpdateSpace() {
	spaceInfo1, err := fsClient.GetSpaceInfo()
	if err != nil {
		fmt.Println("UpdateSpace GetSpaceInfo1 error: ", err.Error())
		return
	}
	common.PrintStruct(*spaceInfo1)

	timeExpired := uint64(time.Now().Unix()) + 3600*24
	txHash, err := fsClient.UpdateSpace(1024*2048, timeExpired)
	if err != nil {
		fmt.Println("UpdateSpace error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(15*time.Second, txHash)

	spaceInfo2, err := fsClient.GetSpaceInfo()
	if err != nil {
		fmt.Println("UpdateSpace GetSpaceInfo2 error: ", err.Error())
		return
	}
	common.PrintStruct(*spaceInfo2)

	if spaceInfo1.Volume != spaceInfo2.Volume || spaceInfo1.TimeExpired != spaceInfo2.TimeExpired {
		fmt.Println("UpdateSpace Success")
	} else {
		fmt.Println("UpdateSpace Failed")
	}
}

func GetGlobalParam() {
	var err error
	globalParam, err = fsClient.GetGlobalParam()
	if err != nil {
		fmt.Printf("APP GetGlobalParam error: %s\n", err.Error())
		return
	}
	common.PrintStruct(*globalParam)
}

func GetNodeInfoList() {
	nodeInfoList, err := fsClient.GetNodeInfoList(9999999)
	if err != nil {
		fmt.Printf("APP GetNodeInfoList error: %s\n", err.Error())
		return
	} else {
		fmt.Printf("NodeInfoListLen: %d\n", len(nodeInfoList.NodesInfo))
		for _, nodeInfo := range nodeInfoList.NodesInfo {
			common.PrintStruct(nodeInfo)
		}
	}
}

func GetFileList() {
	fileHashList, err := fsClient.GetFileList()
	if err != nil {
		fmt.Printf("APP GetFileList error: %s\n", err.Error())
		return
	} else {
		for _, fileHash := range fileHashList.FilesH {
			fmt.Printf("FileHash: %s\n", fmt.Sprintf("%s", fileHash.FHash))
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
		fmt.Println("StoreFile error: ", err.Error())
		return
	}

	if len(storeErrors.ObjectErrors) == 0 {
		fmt.Printf("StoreFile success\n")
		return
	}
	for k, v := range storeErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}

	once.Do(func() {
		if err := connectFs(); err != nil {
			fmt.Println("connectFs error: ", err.Error())
			os.Exit(0)
		}
		fmt.Println("connection success")
	},
	)

	if err = sendToFs("StoreFile" + "|" + TestFileHash); err != nil {
		fmt.Println("sendToFs error: ", err.Error())
		return
	}
	closeConn()
}

func GetFileInfo(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		fmt.Printf("GetFileInfo fileHash: %s error: %s\n", fileHash, err.Error())
		return
	}
	common.PrintStruct(*fileInfo)

	filePdpNeedCount := (fileInfo.TimeExpired-fileInfo.TimeStart)/fileInfo.PdpInterval + 1
	fmt.Printf("TotalPdpNeedCount: %d\n", filePdpNeedCount)
}

func RenewFile(fileHash string) {
	fileInfo, err := fsClient.GetFileInfo(fileHash)
	if err != nil {
		fmt.Printf("RenewFile GetFileInfo fileHash: %s error: %s\n", fileHash, err.Error())
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
		fmt.Println("RenewFile error: ", err.Error())
		return
	}

	if len(renewErrors.ObjectErrors) == 0 {
		fmt.Printf("RenewFiles success\n")
		return
	}
	for k, v := range renewErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}
}

func DeleteFile(fileHash string) {
	_, err, delErrors := fsClient.DeleteFiles([]string{fileHash})
	if err != nil {
		fmt.Println("DeleteFile error: ", err.Error())
		return
	}

	if len(delErrors.ObjectErrors) == 0 {
		fmt.Printf("DeleteFile success\n")
		return
	}
	for k, v := range delErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}

}

func TransferFile(fileHash string, newOwner string) {
	newOwnerAddr, err := utils.AddressFromBase58(newOwner)
	if err != nil {
		fmt.Println("ChangeOwner AddressFromBase58 error: ", err.Error())
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
		fmt.Println("ChangeOwner error: ", err.Error())
		return
	}
	if len(transferErrors.ObjectErrors) == 0 {
		fmt.Printf("TransferFile success\n")
		return
	}
	for k, v := range transferErrors.ObjectErrors  {
		fmt.Printf("%s | %s\n", k, v)
	}
}

func GetPdpInfoList(fileHash string) {
	pdpRecordList, err := fsClient.GetFilePdpRecordList(fileHash)
	if err != nil {
		fmt.Printf("APP GetFilePdpRecordList error: %s\n", err.Error())
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
		fmt.Printf("StoreFile GetFileInfo fileHash: %s error: %s\n", fileHash, err.Error())
		return
	} else if fileInfo == nil {
		fmt.Println("StoreFile GetFileInfo failed, fileInfo is nil")
		return
	}

	pdpRecordList, err := fsClient.GetFilePdpRecordList(fileHash)
	if err != nil {
		fmt.Printf("APP GetFilePdpRecordList error: %s\n", err.Error())
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
		fmt.Println("FileReadPledge failed error: ", err.Error())
		return
	}
	fsClient.PollForTxConfirmed(14*time.Second, readTx)

	haveReadBlockNum := uint64(0)
	readPledge, err := fsClient.GetFileReadPledge(fileHash, fsClient.WalletAddr)
	if err != nil {
		fmt.Println("GetFileReadPledge failed error: ", err.Error())
		return
	}
	for _, readPlan := range readPledge.ReadPlans {
		if readPlan.NodeAddr == fsClient.WalletAddr {
			haveReadBlockNum = readPlan.HaveReadBlockNum
		}
	}

	once.Do(func() {
		if err := connectFs(); err != nil {
			fmt.Println("connectFs error: ", err.Error())
			os.Exit(0)
		}
		fmt.Println("connection success")
	},
	)

	if err = sendToFs("ReadFile" + "|" + fileHash + "|" + fsClient.WalletAddr.ToBase58()); err != nil {
		fmt.Println("sendToFs error: ", err.Error())
		return
	}

	for i := uint64(0); i < fileInfo.FileBlockCount; i++ {
		fileReadSlice, err := fsClient.GenFileReadSettleSlice([]byte(fileHash), readPlans[0].NodeAddr, i+haveReadBlockNum, 1)
		if err != nil {
			fmt.Printf("GenFileReadSettleSlice error: %s\n", err.Error())
			return
		}
		sliceData := common.FileReadSettleSliceSerialize(fileReadSlice)
		sliceString := hex.EncodeToString(sliceData)

		fmt.Println("sendToFs FileReadSettleSlice")
		sendToFs(sliceString)
	}
	closeConn()
}
