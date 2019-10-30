package client

import (
	"fmt"
	"testing"

	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
)

var ontFs *OntFsClient

var txt = "QmevhnWdtmz89BMXuuX5pSY2uZtqKLz7frJsrCojT5kmb6"

func TestMain(m *testing.M) {
	ontFs = Init("./wallet.dat", "pwd", "http://localhost:20336")
	if ontFs == nil {
		fmt.Println("Init error")
	}
	m.Run()
}

func TestOntFsClient_GetNodeInfoList(t *testing.T) {
	ret, err := ontFs.GetNodeInfoList()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	nodeInfoList := (*fs.FsNodeInfoList)(ret)

	for _, nodeInfo := range nodeInfoList.NodesInfo {
		fmt.Println("FsNodeQuery Success")
		fmt.Println("Pledge: ", nodeInfo.Pledge)
		fmt.Println("Profit: ", nodeInfo.Profit)
		fmt.Println("Volume: ", nodeInfo.Volume)
		fmt.Println("RestVol: ", nodeInfo.RestVol)
		fmt.Println("NodeAddr: ", nodeInfo.NodeAddr)
		fmt.Println("NodeNetAddr: ", nodeInfo.NodeNetAddr)
		fmt.Println("ServiceTime:", nodeInfo.ServiceTime)
	}
}

func TestOntFsClient_StoreFile(t *testing.T) {
	proveParam := []byte("ProveProveProveProveProveProveProveProveProveProveProveProve")
	ret, err := ontFs.StoreFile(txt, 12, 1582420196,
		3, []byte("helloWorld.txt"), proveParam, 1, 65536)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println(ret)
}

func TestOntFsClient_GetFileInfo(t *testing.T) {
	fileInfo, err := ontFs.GetFileInfo(txt)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("FileHash:", fileInfo.FileHash)
	fmt.Println("FileOwner:", fileInfo.FileOwner)
	fmt.Println("FileDesc:", fileInfo.FileDesc)
	fmt.Println("CopyNumber:", fileInfo.CopyNumber)

	fmt.Println("PayAmount:", fileInfo.PayAmount)
	fmt.Println("RestAmount:", fileInfo.RestAmount)

	fmt.Println("TimeStart:", fileInfo.TimeStart)
	fmt.Println("TimeExpired:", fileInfo.TimeExpired)

	fmt.Println("FileBlockCount:", fileInfo.FileBlockCount)
	fmt.Println("RealFileSize:", fileInfo.RealFileSize)
	fmt.Println("ValidFlag:", fileInfo.ValidFlag)
	fmt.Println("StorageType:", fileInfo.StorageType)
	fmt.Println("FileProveParam:", string(fileInfo.PdpParam))

}

func TestOntFsClient_GetFilePdpRecordList(t *testing.T) {
	pdpRecordList, err := ontFs.GetFilePdpRecordList(txt)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	for _, pdpRecord := range pdpRecordList.PdpRecords {
		fmt.Println("FileOwner: ", pdpRecord.FileOwner)
		fmt.Println("NodeAddr:", pdpRecord.NodeAddr)
		fmt.Println("FileHash:", pdpRecord.FileHash)
		fmt.Println("PdpCount:", pdpRecord.PdpCount)
		fmt.Println("NextHeight:", pdpRecord.NextHeight)
		fmt.Println("LastPdpTime:", pdpRecord.LastPdpTime)
		fmt.Println("SettleFlag:", pdpRecord.SettleFlag)
	}
}
