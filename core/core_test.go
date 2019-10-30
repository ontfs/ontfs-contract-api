package core

import (
	"fmt"
	"testing"
)

var ontFs *OntFs

var txt = "QmevhnWdtmz89BMXuuX5pSY2uZtqKLz7frJsrCojT5kmb6"

func TestMain(m *testing.M) {
	ontFs = Init("./wallet.dat", "pwd", "http://localhost:20336")
	if ontFs == nil {
		fmt.Println("Init error")
	}
	m.Run()
}
//
//func TestOntFs_OntFsInit(t *testing.T) {
//	ontFs.OntFsInit(2000, 2, 1, 1, 32, 120, 1024*1024)
//}

func TestOntFs_GetGlobalParam(t *testing.T) {
	globalParam, err := ontFs.GetGlobalParam()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("GetGlobalParam Success")
	fmt.Println("GasPerKbForRead:", globalParam.GasPerKbForRead,
		"NodePerKbPledge:", globalParam.NodePerKbPledge,
		"NodeMinVolume:", globalParam.NodeMinVolume,
		"MinDownLoadFee:", globalParam.MinDownLoadFee,
		"GasPerKbForSaveWithFile:", globalParam.GasPerKbForSaveWithFile,
		"GasPerKbForSaveWithSpace:", globalParam.GasPerKbForSaveWithSpace)
}

func TestOntFs_NodeRegister(t *testing.T) {
	_, err := ontFs.NodeRegister(1000, 10000, "https://127.0.0.1:5002")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("NodeRegister Success")
}

func TestOntFs_NodeQuery(t *testing.T) {
	nodeInfo, err := ontFs.NodeQuery(ontFs.WalletAddr)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("NodeQuery Success")
	fmt.Println("NodeAddr: ", nodeInfo.NodeAddr.ToBase58())
	fmt.Println("Pledge: ", nodeInfo.Pledge)
	fmt.Println("Profit: ", nodeInfo.Profit)
	fmt.Println("Volume: ", nodeInfo.Volume)
	fmt.Println("RestVol: ", nodeInfo.RestVol)
	fmt.Println("NodeNetAddr: ", nodeInfo.NodeNetAddr)
	fmt.Println("ServiceTime:", nodeInfo.ServiceTime)
}

func TestOntFs_NodeUpdate(t *testing.T) {
	_, err := ontFs.NodeUpdate(100000, 100000, "https://127.0.0.1:5001")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("NodeUpdate Success")
}

func TestOntFs_NodeCancel(t *testing.T) {
	_, err := ontFs.NodeCancel()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("NodeCancel Success")
}

func TestOntFs_NodeWithDrawProfit(t *testing.T) {
	_, err := ontFs.NodeWithDrawProfit()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("NodeWithDrawProfit Success")
}

func TestOntFs_FileProve(t *testing.T) {
	_, err := ontFs.FileProve(txt, nil, "", 1)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println("TestOntFs_FileProve Success")
}
