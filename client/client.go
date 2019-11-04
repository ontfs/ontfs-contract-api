package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	apiComm "github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontology-crypto/keypair"
	ontSdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
	ontUtils "github.com/ontio/ontology/smartcontract/service/native/utils"
)

type OntFsClient struct {
	WalletPath    string
	Password      []byte
	WalletAddr    common.Address
	GasPrice      uint64
	GasLimit      uint64
	OntSdk        *ontSdk.OntologySdk
	Wallet        *ontSdk.Wallet
	DefAcc        *ontSdk.Account
	OntRpcSrvAddr string
}

const contractVersion = byte(0)

var contractAddr common.Address

func Init(walletPath string, walletPwd string, ontRpcSrvAddr string) *OntFsClient {
	contractAddr = ontUtils.OntFSContractAddress
	ontFs := &OntFsClient{
		WalletPath:    walletPath,
		Password:      []byte(walletPwd),
		GasPrice:      uint64(0),
		GasLimit:      uint64(20000),
		OntRpcSrvAddr: ontRpcSrvAddr,
	}

	ontFs.OntSdk = ontSdk.NewOntologySdk()
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

func (c *OntFsClient) GetGlobalParam() (*fs.FsGlobalParam, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
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

func (c *OntFsClient) GetNodeInfoList() (*fs.FsNodeInfoList, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(
		contractAddr, contractVersion, fs.FS_GET_NODE_LIST, []interface{}{},
	)
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetNodeInfoList result toByteArray: %s", err.Error())
	}

	var nodeInfoList fs.FsNodeInfoList
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = nodeInfoList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetNodeInfoList Deserialization: %s", err.Error())
		}
		return &nodeInfoList, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *OntFsClient) StoreFile(fileHash string, fileBlockCount uint64, timeExpired uint64, copyNum uint64,
	fileDesc []byte, pdpParam []byte, storageType uint64, realFileSize uint64) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	fileInfo := fs.FileInfo{
		FileHash:       []byte(fileHash),
		FileOwner:      c.DefAcc.Address,
		FileDesc:       fileDesc,
		FileBlockCount: fileBlockCount,
		RealFileSize:   realFileSize,
		CopyNumber:     copyNum,
		TimeExpired:    timeExpired,
		PdpParam:       pdpParam,
		StorageType:    storageType,
	}
	fileInfoList := fs.FileInfoList{}
	fileInfoList.FilesI = append(fileInfoList.FilesI, fileInfo)

	sink := common.NewZeroCopySink(nil)
	fileInfoList.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_STORE_FILES, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *OntFsClient) RenewFile(fileHashStr string, renewTimes uint64) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	fileRenew := fs.FileReNew{
		FileHash:       []byte(fileHashStr),
		FileOwner:      c.WalletAddr,
		Payer:          c.WalletAddr,
		NewTimeExpired: renewTimes,
	}
	fileReNewList := fs.FileReNewList{}
	fileReNewList.FilesReNew = append(fileReNewList.FilesReNew, fileRenew)

	sink := common.NewZeroCopySink(nil)
	fileReNewList.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc,
		contractVersion, contractAddr, fs.FS_RENEW_FILES, []interface{}{sink.Bytes()},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *OntFsClient) DeleteFiles(fileHashStrs []string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	var fileDelList fs.FileDelList
	for _, fileHashStr := range fileHashStrs {
		fileDelList.FilesDel = append(fileDelList.FilesDel, fs.FileDel{FileHash:[]byte(fileHashStr)})
	}

	sink := common.NewZeroCopySink(nil)
	fileDelList.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_DELETE_FILES, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *OntFsClient) GetFileInfo(fileHashStr string) (*fs.FileInfo, error) {
	fileHash := []byte(fileHashStr)
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
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

func (c *OntFsClient) ChangeFileOwner(fileHashStr string, newOwner common.Address) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileTransfer := fs.FileTransfer{
		FileHash: []byte(fileHashStr),
		OriOwner: c.DefAcc.Address,
		NewOwner: newOwner,
	}
	fileTransferList := fs.FileTransferList{}
	fileTransferList.FilesTransfer = append(fileTransferList.FilesTransfer, fileTransfer)

	sink := common.NewZeroCopySink(nil)
	fileTransferList.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_TRANSFER_FILES, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *OntFsClient) GetFileList() (*fs.FileHashList, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_FILE_LIST, []interface{}{c.WalletAddr})
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetFileList result toByteArray: %s", err.Error())
	}

	var fileList fs.FileHashList
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := common.NewZeroCopySource(retInfo.Info)
		if err = fileList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFileList error: %s", err.Error())
		}
		return &fileList, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *OntFsClient) GetFilePdpRecordList(fileHashStr string) (*fs.PdpRecordList, error) {
	fileHash := []byte(fileHashStr)
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
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

//
//func (c *OntFsClient) WhiteListOp(fileHashStr string, op uint64, whiteList fs.WhiteList) ([]byte, error) {
//	if c.DefAcc == nil {
//		return nil, errors.New("DefAcc is nil")
//	}
//	if op != fs.ADD && op != fs.ADD_COV && op != fs.DEL && op != fs.DEL_ALL {
//		return nil, errors.New("Param [op] error")
//	}
//	fileHash := []byte(fileHashStr)
//	whiteListOp := fs.WhiteListOp{FileHash: fileHash, Op: op, List: whiteList}
//	buf := new(bytes.Buffer)
//	if err := whiteListOp.Serialize(buf); err != nil {
//		return nil, fmt.Errorf("WhiteListOp serialize error: %s", err.Error())
//	}
//	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc,
//		contractVersion, contractAddr, fs.FS_WHITE_LIST_OP, []interface{}{buf.Bytes()},
//	)
//	if err != nil {
//		return nil, err
//	}
//	return ret.ToArray(), err
//}
//
//func (c *OntFsClient) GetWhiteList(fileHashStr string) (*fs.WhiteList, error) {
//	fileHash := []byte(fileHashStr)
//	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
//		fs.FS_GET_WHITE_LIST, []interface{}{fileHash},
//	)
//	if err != nil {
//		return nil, err
//	}
//	data, err := ret.Result.ToByteArray()
//	if err != nil {
//		return nil, fmt.Errorf("GetProveDetails result toByteArray: %s", err.Error())
//	}
//	var whiteList fs.WhiteList
//	retInfo := fs.DecRet(data)
//	if retInfo.Ret {
//		whiteListReader := bytes.NewReader(retInfo.Info)
//		err = whiteList.Deserialize(whiteListReader)
//		if err != nil {
//			return nil, fmt.Errorf("GetWhiteList deserialize error: %s", err.Error())
//		}
//		return &whiteList, err
//	} else {
//		return nil, errors.New(string(retInfo.Info))
//	}
//}

func (c *OntFsClient) FileReadPledge(fileHashStr string, readPlans []fs.ReadPlan) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileHash := []byte(fileHashStr)

	fileReadPledge := &fs.ReadPledge{
		FileHash:     fileHash,
		Downloader:   c.DefAcc.Address,
		BlockHeight:  0,
		ExpireHeight: 0,
		RestMoney:    0,
		ReadPlans:    readPlans,
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_READ_FILE_PLEDGE, []interface{}{fileReadPledge})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *OntFsClient) GetFileReadPledge(fileHashStr string, downloader common.Address) (*fs.ReadPledge, error) {
	fileHash := []byte(fileHashStr)
	getReadPledge := &fs.GetReadPledge{
		FileHash:   fileHash,
		Downloader: downloader,
	}
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
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

func (c *OntFsClient) GenFileReadSettleSlice(fileHash []byte, payTo common.Address, sliceId uint64) (*fs.FileReadSettleSlice, error) {
	settleSlice := fs.FileReadSettleSlice{
		FileHash: fileHash,
		PayFrom:  c.DefAcc.Address,
		PayTo:    payTo,
		SliceId:  sliceId,
	}
	sink := common.NewZeroCopySink(nil)
	settleSlice.Serialization(sink)

	signData, err := apiComm.Sign(c.DefAcc, sink.Bytes())
	if err != nil {
		return nil, fmt.Errorf("FileReadSettleSlice Sign error: %s", err.Error())
	}
	settleSlice.Sig = signData
	settleSlice.PubKey = keypair.SerializePublicKey(c.DefAcc.PublicKey)
	return &settleSlice, nil
}

func (c *OntFsClient) CancelFileRead(fileHashStr string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileHash := []byte(fileHashStr)
	getReadPledge := &fs.GetReadPledge{
		FileHash:   fileHash,
		Downloader: c.DefAcc.Address,
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_CANCEL_FILE_READ, []interface{}{getReadPledge})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *OntFsClient) PollForTxConfirmed(timeout time.Duration, txHash []byte) (bool, error) {
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
		ret, err := c.OntSdk.GetBlockHeightByTxHash(txHashStr)
		if err != nil || ret == 0 {
			continue
		}
		return true, nil
	}
	return false, fmt.Errorf("timeout after %d (s)", secs)
}
