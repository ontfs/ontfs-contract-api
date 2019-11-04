package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/pdp"
	"github.com/ontio/ontology-crypto/signature"
	ont "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const contractVersion = byte(0)

var contractAddr common.Address

type OntFs struct {
	WalletPath    string
	Password      []byte
	WalletAddr    common.Address
	GasPrice      uint64
	GasLimit      uint64
	OntSdk        *ont.OntologySdk
	Wallet        *ont.Wallet
	DefAcc        *ont.Account
	OntRpcSrvAddr string
}

func Init(walletPath string, walletPwd string, ontRpcSrvAddr string) *OntFs {
	contractAddr = utils.OntFSContractAddress
	ontFs := &OntFs{
		WalletPath:    walletPath,
		Password:      []byte(walletPwd),
		GasPrice:      uint64(0),
		GasLimit:      uint64(20000),
		OntRpcSrvAddr: ontRpcSrvAddr,
	}

	ontFs.OntSdk = ont.NewOntologySdk()
	ontFs.OntSdk.NewRpcClient().SetAddress(ontFs.OntRpcSrvAddr)

	if len(walletPath) != 0 {
		var err error
		ontFs.Wallet, err = ontFs.OntSdk.OpenWallet(ontFs.WalletPath)
		if err != nil {
			fmt.Printf("Account.Open error:%s\n", err)
			return nil
		}
		ontFs.DefAcc, err = ontFs.Wallet.GetDefaultAccount(ontFs.Password)
		if err != nil {
			fmt.Printf("GetDefaultAccount error:%s\n", err)
			return nil
		}
		ontFs.WalletAddr = ontFs.DefAcc.Address
	} else {
		ontFs.Wallet = nil
		ontFs.DefAcc = nil
	}
	return ontFs
}

//func (d *OntFs) OntFsInit(fsGasPrice, gasPerKBPerBlock, gasPerKBForRead, gasForChallenge,
//	maxProveBlockNum, minChallengeRate uint64, minVolume uint64) ([]byte, error) {
//	if d.DefAcc == nil {
//		return nil, errors.New("DefAcc is nil")
//	}
//	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc,
//		contractVersion, contractAddr, fs.FS_INIT,
//		[]interface{}{&fs.FsSetting{FsGasPrice: fsGasPrice,
//			GasPerKBPerBlock: gasPerKBPerBlock,
//			GasPerKBForRead:  gasPerKBForRead,
//			GasForChallenge:  gasForChallenge,
//			MaxProveBlockNum: maxProveBlockNum,
//			MinChallengeRate: minChallengeRate,
//			MinVolume:        minVolume}},
//	)
//	if err != nil {
//		return nil, err
//	}
//	return ret.ToArray(), err
//}

func (d *OntFs) GetGlobalParam() (*fs.FsGlobalParam, error) {
	ret, err := d.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_GLOBAL_PARAM, []interface{}{})
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetGlobalParam result toByteArray: %s", err.Error())
	}

	var globalParam fs.FsGlobalParam
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = globalParam.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetGlobalParam error: %s", err.Error())
		}
		return &globalParam, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (d *OntFs) NodeRegister(volume uint64, serviceTime uint64, nodeNetAddr string) ([]byte, error) {
	if d.DefAcc == nil {
		return nil, errors.New("NodeRegister DefAcc is nil")
	}
	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_REGISTER, []interface{}{&fs.FsNodeInfo{Pledge: 0, Profit: 0, Volume: volume, RestVol: 0,
			ServiceTime: serviceTime, NodeAddr: d.WalletAddr, NodeNetAddr: []byte(nodeNetAddr)}})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (d *OntFs) NodeQuery(nodeWallet common.Address) (*fs.FsNodeInfo, error) {
	ret, err := d.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_NODE_QUERY, []interface{}{nodeWallet})
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("NodeQuery result toByteArray: %s", err.Error())
	}

	var fsNodeInfo fs.FsNodeInfo
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = fsNodeInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("NodeQuery error: %s", err.Error())
		}
		return &fsNodeInfo, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (d *OntFs) NodeUpdate(volume uint64, serviceTime uint64, nodeNetAddr string) ([]byte, error) {
	if d.DefAcc == nil {
		return nil, errors.New("NodeUpdate DefAcc is nil")
	}
	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_UPDATE, []interface{}{&fs.FsNodeInfo{Pledge: 0, Profit: 0, Volume: volume, RestVol: 0,
			ServiceTime: serviceTime, NodeAddr: d.WalletAddr, NodeNetAddr: []byte(nodeNetAddr)}},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (d *OntFs) NodeCancel() ([]byte, error) {
	if d.DefAcc == nil {
		return nil, errors.New("NodeCancel DefAcc is nil")
	}
	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_CANCEL, []interface{}{d.WalletAddr})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (d *OntFs) NodeWithDrawProfit() ([]byte, error) {
	if d.DefAcc == nil {
		return nil, errors.New("NodeWithDrawProfit DefAcc is nil")
	}
	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_WITH_DRAW_PROFIT, []interface{}{d.WalletAddr},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (d *OntFs) GetFileInfo(fileHashStr string) (*fs.FileInfo, error) {
	fileHash := []byte(fileHashStr)
	ret, err := d.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_FILE_INFO, []interface{}{fileHash},
	)
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetFileInfo result toByteArray: %s", err.Error())
	}

	var fileInfo fs.FileInfo
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = fileInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFileInfo error: %s", err.Error())
		}
		return &fileInfo, err
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (d *OntFs) FileProve(fileHashStr string, multiRes []byte, addResStr string, blockHeight uint64) ([]byte, error) {
	if d.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileHash := []byte(fileHashStr)
	addRes := []byte(addResStr)
	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc, contractVersion, contractAddr,
		fs.FS_FILE_PROVE, []interface{}{&fs.PdpData{
			FileHash:        fileHash,
			NodeAddr:        d.WalletAddr,
			MultiRes:        multiRes,
			AddRes:          addRes,
			ChallengeHeight: blockHeight,
		}},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (d *OntFs) GetFileReadPledge(fileHashStr string, downloader common.Address) (*fs.ReadPledge, error) {
	fileHash := []byte(fileHashStr)
	getReadPledge := &fs.GetReadPledge{
		FileHash:   fileHash,
		Downloader: downloader,
	}
	ret, err := d.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_READ_PLEDGE, []interface{}{getReadPledge})
	if err != nil {
		return nil, err
	}

	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetFileReadPledge result toByteArray: %s", err.Error())
	}

	var readPledge fs.ReadPledge
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = readPledge.Deserialization(src); err != nil {
			return nil, err
		}
		return &readPledge, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (d *OntFs) FileReadProfitSettle(fileReadSettleSlice fs.FileReadSettleSlice) ([]byte, error) {
	if d.DefAcc == nil {
		return nil, errors.New("FileReadProfitSettle DefAcc is nil")
	}
	ret, err := d.OntSdk.Native.InvokeNativeContract(d.GasPrice, d.GasLimit, d.DefAcc, contractVersion, contractAddr,
		fs.FS_READ_FILE_SETTLE, []interface{}{&fileReadSettleSlice},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (d *OntFs) VerifyFileReadSettleSlice(settleSlice fs.FileReadSettleSlice) (bool, error) {
	tmpSettleSlice := fs.FileReadSettleSlice{
		FileHash: settleSlice.FileHash,
		PayFrom:  settleSlice.PayFrom,
		PayTo:    settleSlice.PayTo,
		SliceId:  settleSlice.SliceId,
	}
	sink := common.NewZeroCopySink(nil)
	tmpSettleSlice.Serialization(sink)

	signValue, err := signature.Deserialize(settleSlice.Sig)
	if err != nil {
		return false, fmt.Errorf("FileReadSettleSlice signature deserialize error: %s", err.Error())
	}
	pubKey, err := keypair.DeserializePublicKey(settleSlice.PubKey)
	if err != nil {
		return false, fmt.Errorf("FileReadSettleSlice deserialize PublicKey( error: %s", err.Error())
	}
	result := signature.Verify(pubKey, sink.Bytes(), signValue)
	return result, nil
}

func (d *OntFs) GetFilePdpRecordList(fileHashStr string) (*fs.PdpRecordList, error) {
	fileHash := []byte(fileHashStr)
	ret, err := d.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_PDP_INFO_LIST, []interface{}{fileHash},
	)
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetFilePdpRecordList result toByteArray: %s", err.Error())
	}

	var pdpRecordList fs.PdpRecordList
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = pdpRecordList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFilePdpRecordList deserialize error: %s", err.Error())
		}
		return &pdpRecordList, err
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (d *OntFs) GenChallenge(nodeAddr common.Address, hash []byte, fileBlockNum, proveNum uint64) []pdp.Challenge {
	return fs.GenChallenge(nodeAddr, hash, uint32(fileBlockNum), uint32(proveNum))
}

func (d *OntFs) PollForTxConfirmed(timeout time.Duration, txHash []byte) (bool, error) {
	if len(txHash) == 0 {
		return false, fmt.Errorf("txHash is empty")
	}
	txHashStr := hex.EncodeToString(common.ToArrayReverse(txHash))
	secs := int(timeout / time.Second)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		time.Sleep(time.Second)
		ret, err := d.OntSdk.GetBlockHeightByTxHash(txHashStr)
		if err != nil || ret == 0 {
			continue
		}
		return true, nil
	}
	return false, fmt.Errorf("timeout after %d (s)", secs)
}
