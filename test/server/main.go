package main

import (
	"flag"
	"time"

	"github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontfs-contract-api/core"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"
)

const TestFileHash = "FileTest"

var fsCore *core.Core
var globalParam *ontfs.FsGlobalParam

var action = struct {
	getGlobalParam bool
	nodeRegister   bool
	nodeQuery      bool
	nodeUpdate     bool
	nodeCancel     bool
	withdrawProfit bool
	getFileInfo    bool
	fileHash       string
}{}

func main() {
	flag.BoolVar(&action.getGlobalParam, "getGlobalParam", false, "getGlobalParam")
	flag.BoolVar(&action.nodeRegister, "nodeRegister", false, "nodeRegister")
	flag.BoolVar(&action.nodeQuery, "nodeQuery", false, "nodeQuery")
	flag.BoolVar(&action.nodeUpdate, "nodeUpdate", false, "nodeUpdate")
	flag.BoolVar(&action.nodeCancel, "nodeCancel", false, "nodeCancel")
	flag.BoolVar(&action.withdrawProfit, "withdraw", false, "withdrawProfit")
	flag.BoolVar(&action.getFileInfo, "getFileInfo", false, "getFileInfo")
	flag.StringVar(&action.fileHash, "fileHash", TestFileHash, "fileHash")
	flag.Parse()

	fsCore = core.Init("./wallet.dat", "pwd", "http://localhost:33894", 0, 20000)
	if fsCore == nil {
		log.Error("fsNode Init error")
		return
	}

	if action.getGlobalParam {
		GetGlobalParam()
	} else if action.nodeRegister {
		RegisterNode()
	} else if action.nodeUpdate {
		UpdateNode()
	} else if action.nodeQuery {
		QueryNode()
	} else if action.nodeCancel {
		CancelNode()
	} else if action.withdrawProfit {
		WithDrawProfit()
	} else if action.getFileInfo {
		GetFileInfo(action.fileHash)
	} else {
		FsServer()
	}
}

func GetGlobalParam() {
	var err error
	globalParam, err = fsCore.GetGlobalParam()
	if err != nil {
		log.Error("GetGlobalParam error")
		return
	}
	common.PrintStruct(*globalParam)
}

func RegisterNode() {
	serviceDueTime := time.Now().Unix() + 100000
	_, err := fsCore.NodeRegister(1024*1024*1024, uint64(serviceDueTime), 4*60*60, "tcp://10.0.1.66:3389")
	if err != nil {
		log.Errorf("NodeRegister error: %s", err.Error())
		return
	}
}

func QueryNode() {
	nodeInfo, err := fsCore.NodeQuery(fsCore.WalletAddr)
	if err != nil {
		log.Errorf("NodeQuery error: %s", err.Error())
		return
	} else {
		common.PrintStruct(*nodeInfo)
	}
}

func UpdateNode() {
	serviceDueTime := time.Now().Unix() + 100000
	_, err := fsCore.NodeUpdate(1024*1024*1024, uint64(serviceDueTime), 4*60*60, "tcp://10.0.1.66:1004")
	if err != nil {
		log.Errorf("NodeUpdate error: %s", err.Error())
		return
	}
}

func CancelNode() {
	_, err := fsCore.NodeCancel()
	if err != nil {
		log.Errorf("NodeCancel error: %s", err.Error())
		return
	}
}

func WithDrawProfit() {
	_, err := fsCore.NodeWithDrawProfit()
	if err != nil {
		log.Errorf("NodeWithDrawProfit error: %s", err.Error())
		return
	}
}

func GetFileInfo(fileHash string) {
	fileInfo, err := fsCore.GetFileInfo(fileHash)
	if err != nil {
		log.Errorf("GetFileInfo fileHash: %s error: %s", fileHash, err.Error())
		return
	}
	common.PrintStruct(*fileInfo)
}
