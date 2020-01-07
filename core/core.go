package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	ont "github.com/ontio/ontology-go-sdk"
	ccom "github.com/ontio/ontology/common"
	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"strings"
)

const contractVersion = byte(0)
const defaultMinPdpInterval = uint64(10 * 60)

var contractAddr ccom.Address
var contractAddrStr string

type Core struct {
	WalletPath    string
	Password      []byte
	WalletAddr    ccom.Address
	GasPrice      uint64
	GasLimit      uint64
	OntSdk        *ont.OntologySdk
	Wallet        *ont.Wallet
	DefAcc        *ont.Account
	OntRpcSrvAddr string
}

func Init(walletPath string, walletPwd string, ontRpcSrvAddr string, gasPrice uint64, gasLimit uint64) *Core {
	contractAddr = utils.OntFSContractAddress
	contractAddrStr = contractAddr.ToHexString()

	ontFs := &Core{
		WalletPath:    walletPath,
		Password:      []byte(walletPwd),
		GasPrice:      gasPrice,
		GasLimit:      gasLimit,
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

func (c *Core) GetGlobalParam() (*fs.FsGlobalParam, error) {
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = globalParam.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetGlobalParam error: %s", err.Error())
		}
		return &globalParam, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) NodeRegister(volume uint64, serviceTime uint64, minPdpInterval uint64, nodeNetAddr string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeRegister DefAcc is nil")
	}
	fsNodeInfo := fs.FsNodeInfo{
		Pledge:         0,
		Profit:         0,
		Volume:         volume,
		RestVol:        0,
		ServiceTime:    serviceTime,
		MinPdpInterval: minPdpInterval,
		NodeAddr:       c.WalletAddr,
		NodeNetAddr:    []byte(nodeNetAddr),
	}

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_REGISTER, []interface{}{&fsNodeInfo})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("NodeRegister tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) NodeQuery(nodeWallet ccom.Address) (*fs.FsNodeInfo, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = fsNodeInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("NodeQuery error: %s", err.Error())
		}
		return &fsNodeInfo, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) NodeUpdate(volume uint64, serviceTime uint64, minPdpInterval uint64, nodeNetAddr string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeUpdate DefAcc is nil")
	}
	fsNodeInfo := fs.FsNodeInfo{
		Pledge:         0,
		Profit:         0,
		Volume:         volume,
		RestVol:        0,
		ServiceTime:    serviceTime,
		MinPdpInterval: minPdpInterval,
		NodeAddr:       c.WalletAddr,
		NodeNetAddr:    []byte(nodeNetAddr),
	}

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_UPDATE, []interface{}{&fsNodeInfo},
	)
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("NodeUpdate tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) NodeCancel() ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeCancel DefAcc is nil")
	}
	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_CANCEL, []interface{}{c.WalletAddr})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("NodeCancel tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) NodeWithDrawProfit() ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeWithDrawProfit DefAcc is nil")
	}
	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_WITH_DRAW_PROFIT, []interface{}{c.WalletAddr},
	)
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("NodeWithDrawProfit tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) FileProve(fileHashStr string, proveData []byte, blockHeight uint64) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileHash := []byte(fileHashStr)
	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_FILE_PROVE, []interface{}{&fs.PdpData{
			FileHash:        fileHash,
			NodeAddr:        c.WalletAddr,
			ProveData:       proveData,
			ChallengeHeight: blockHeight,
		}},
	)
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("FileProve tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) GetFileReadPledge(fileHashStr string, downloader ccom.Address) (*fs.ReadPledge, error) {
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = readPledge.Deserialization(src); err != nil {
			return nil, err
		}
		return &readPledge, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) FileReadProfitSettle(fileReadSettleSlice *fs.FileReadSettleSlice) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("FileReadProfitSettle DefAcc is nil")
	}
	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_READ_FILE_SETTLE, []interface{}{fileReadSettleSlice},
	)
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("FileReadProfitSettle tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) VerifyFileReadSettleSlice(settleSlice *fs.FileReadSettleSlice) (bool, error) {
	tmpSettleSlice := fs.FileReadSettleSlice{
		FileHash:     settleSlice.FileHash,
		PayFrom:      settleSlice.PayFrom,
		PayTo:        settleSlice.PayTo,
		SliceId:      settleSlice.SliceId,
		PledgeHeight: settleSlice.PledgeHeight,
	}
	sink := ccom.NewZeroCopySink(nil)
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

func (c *Core) GetFilePdpRecordList(fileHashStr string) (*fs.PdpRecordList, error) {
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = pdpRecordList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFilePdpRecordList deserialize error: %s", err.Error())
		}
		return &pdpRecordList, err
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) GetNodeInfo(nodeWallet ccom.Address) (*fs.FsNodeInfo, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_NODE_QUERY, []interface{}{nodeWallet})
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetNodeInfo result toByteArray: %s", err.Error())
	}

	var fsNodeInfo fs.FsNodeInfo
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = fsNodeInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetNodeInfo error: %s", err.Error())
		}
		return &fsNodeInfo, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) GetNodeInfoList(count uint64) (*fs.FsNodeInfoList, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_NODE_LIST, []interface{}{count})
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = nodeInfoList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetNodeInfoList Deserialization: %s", err.Error())
		}
		return &nodeInfoList, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) CreateSpace(volume uint64, copyNumber uint64, pdpInterval uint64, timeExpired uint64) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	if pdpInterval < defaultMinPdpInterval {
		return nil, errors.New("pdpInterval value is too small")
	}

	spaceInfo := fs.SpaceInfo{
		SpaceOwner:  c.DefAcc.Address,
		Volume:      volume,
		CopyNumber:  copyNumber,
		PdpInterval: pdpInterval,
		TimeExpired: timeExpired,
	}

	sink := ccom.NewZeroCopySink(nil)
	spaceInfo.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion,
		contractAddr, fs.FS_CREATE_SPACE, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("CreateSpace tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) GetSpaceInfo() (*fs.SpaceInfo, error) {
	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_SPACE_INFO, []interface{}{c.WalletAddr})
	if err != nil {
		return nil, err
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("GetNodeInfoList result toByteArray: %s", err.Error())
	}

	var spaceInfo fs.SpaceInfo
	retInfo := fs.DecRet(data)
	if retInfo.Ret {
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = spaceInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetSpaceInfo Deserialization: %s", err.Error())
		}
		return &spaceInfo, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) UpdateSpace(volume uint64, timeExpired uint64) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	spaceUpdate := fs.SpaceUpdate{
		SpaceOwner:     c.DefAcc.Address,
		Payer:          c.DefAcc.Address,
		NewVolume:      volume,
		NewTimeExpired: timeExpired,
	}

	sink := ccom.NewZeroCopySink(nil)
	spaceUpdate.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion,
		contractAddr, fs.FS_UPDATE_SPACE, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("UpdateSpace tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) DeleteSpace() ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_DELETE_SPACE, []interface{}{c.DefAcc.Address})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("DeleteSpace tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) GetFileList() (*fs.FileHashList, error) {
	height, err := c.OntSdk.GetCurrentBlockHeight()
	if err != nil {
		return nil, fmt.Errorf("GenPassport GetCurrentBlockHeight error: %s", err.Error())
	}

	blockHash, err := c.OntSdk.GetBlockHash(height)
	if err != nil {
		return nil, fmt.Errorf("GenPassport GetBlockHash error: %s", err.Error())
	}

	passport, err := c.GenPassport(height, blockHash.ToArray())
	if err != nil {
		return nil, fmt.Errorf("GetFileList genPassport error: %s", err.Error())
	}

	ret, err := c.OntSdk.Native.PreExecInvokeNativeContract(contractAddr, contractVersion,
		fs.FS_GET_FILE_LIST, []interface{}{passport})
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = fileList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFileList error: %s", err.Error())
		}
		return &fileList, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) GetFileInfo(fileHashStr string) (*fs.FileInfo, error) {
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
		src := ccom.NewZeroCopySource(retInfo.Info)
		if err = fileInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFileInfo error: %s", err.Error())
		}
		return &fileInfo, err
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) StoreFiles(filesInfo []common.FileStore) ([]byte, error, *fs.Errors) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil"), nil
	}

	fileInfoList := fs.FileInfoList{}
	for _, fileInfo := range filesInfo {
		if fileInfo.PdpInterval < defaultMinPdpInterval {
			return nil, errors.New("pdpInterval value is too small"), nil
		}
		fsFileInfo := fs.FileInfo{
			FileHash:       []byte(fileInfo.FileHash),
			FileOwner:      c.DefAcc.Address,
			FileDesc:       []byte(fileInfo.FileDesc),
			FileBlockCount: fileInfo.FileBlockCount,
			RealFileSize:   fileInfo.RealFileSize,
			CopyNumber:     fileInfo.CopyNumber,
			PdpInterval:    fileInfo.PdpInterval,
			TimeExpired:    fileInfo.TimeExpired,
			PdpParam:       fileInfo.PdpParam,
			StorageType:    fileInfo.StorageType,
		}
		fileInfoList.FilesI = append(fileInfoList.FilesI, fsFileInfo)
	}

	sink := ccom.NewZeroCopySink(nil)
	fileInfoList.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion,
		contractAddr, fs.FS_STORE_FILES, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err, nil
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("StoreFiles tx is not confirmed"), nil
	}

	event, err := c.OntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		return txHash.ToArray(), err, nil
	}

	var errorData string
	for _, notify := range event.Notify  {
		if 0 == strings.Compare(contractAddrStr, notify.ContractAddress) {
			errorData = notify.States.(string)
		}
	}

	var objErrors fs.Errors
	if len(errorData) == 0 {
		err := fmt.Errorf("GetSmartContractEvent error")
		return txHash.ToArray(), err, nil
	}
	err = objErrors.FromString(errorData)
	return txHash.ToArray(), err, &objErrors
}

func (c *Core) TransferFiles(fileTransfers []common.FileTransfer) ([]byte, error, *fs.Errors) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil"), nil
	}

	fileTransferList := fs.FileTransferList{}
	for _, fileRenew := range fileTransfers {
		fsFileTransfer := fs.FileTransfer{
			FileHash: []byte(fileRenew.FileHash),
			OriOwner: c.DefAcc.Address,
			NewOwner: fileRenew.NewOwner,
		}
		fileTransferList.FilesTransfer = append(fileTransferList.FilesTransfer, fsFileTransfer)
	}

	sink := ccom.NewZeroCopySink(nil)
	fileTransferList.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_TRANSFER_FILES, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err, nil
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("TransferFiles tx is not confirmed"), nil
	}

	event, err := c.OntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		return txHash.ToArray(), err, nil
	}

	var errorData string
	for _, notify := range event.Notify  {
		if 0 == strings.Compare(contractAddrStr, notify.ContractAddress) {
			errorData = notify.States.(string)
		}
	}

	var objErrors fs.Errors
	if len(errorData) == 0 {
		err := fmt.Errorf("GetSmartContractEvent error")
		return txHash.ToArray(), err, nil
	}
	err = objErrors.FromString(errorData)
	return txHash.ToArray(), err, &objErrors
}

func (c *Core) RenewFiles(filesRenew []common.FileRenew) ([]byte, error, *fs.Errors) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil"), nil
	}

	fileReNewList := fs.FileReNewList{}
	for _, fileRenew := range filesRenew {
		fsFileRenew := fs.FileReNew{
			FileHash:       []byte(fileRenew.FileHash),
			FileOwner:      c.WalletAddr,
			Payer:          c.WalletAddr,
			NewTimeExpired: fileRenew.RenewTime,
		}
		fileReNewList.FilesReNew = append(fileReNewList.FilesReNew, fsFileRenew)
	}

	sink := ccom.NewZeroCopySink(nil)
	fileReNewList.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc,
		contractVersion, contractAddr, fs.FS_RENEW_FILES, []interface{}{sink.Bytes()})

	if err != nil {
		return nil, err, nil
	}
	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("RenewFiles tx is not confirmed"), nil
	}

	event, err := c.OntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		return txHash.ToArray(), err, nil
	}

	var errorData string
	for _, notify := range event.Notify  {
		if 0 == strings.Compare(contractAddrStr, notify.ContractAddress) {
			errorData = notify.States.(string)
		}
	}

	var objErrors fs.Errors
	if len(errorData) == 0 {
		err := fmt.Errorf("GetSmartContractEvent error")
		return txHash.ToArray(), err, nil
	}
	err = objErrors.FromString(errorData)
	return txHash.ToArray(), err, &objErrors
}

func (c *Core) DeleteFiles(fileHashes []string) ([]byte, error, *fs.Errors) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil"), nil
	}

	var fileDelList fs.FileDelList
	for _, fileHashStr := range fileHashes {
		fileDelList.FilesDel = append(fileDelList.FilesDel, fs.FileDel{FileHash: []byte(fileHashStr)})
	}

	sink := ccom.NewZeroCopySink(nil)
	fileDelList.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_DELETE_FILES, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err, nil
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("DeleteFiles tx is not confirmed"), nil
	}

	event, err := c.OntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		return txHash.ToArray(), err, nil
	}

	var errorData string
	for _, notify := range event.Notify  {
		if 0 == strings.Compare(contractAddrStr, notify.ContractAddress) {
			errorData = notify.States.(string)
		}
	}

	var objErrors fs.Errors
	if len(errorData) == 0 {
		err := fmt.Errorf("GetSmartContractEvent error")
		return txHash.ToArray(), err, nil
	}
	err = objErrors.FromString(errorData)
	return txHash.ToArray(), err, &objErrors
}

func (c *Core) FileReadPledge(fileHashStr string, readPlans []fs.ReadPlan) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	fileReadPledge := &fs.ReadPledge{
		FileHash:     []byte(fileHashStr),
		Downloader:   c.DefAcc.Address,
		BlockHeight:  0,
		ExpireHeight: 0,
		RestMoney:    0,
		ReadPlans:    readPlans,
	}

	sink := ccom.NewZeroCopySink(nil)
	fileReadPledge.Serialization(sink)

	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_READ_FILE_PLEDGE, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("FileReadPledge tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) CancelFileRead(fileHashStr string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileHash := []byte(fileHashStr)
	getReadPledge := &fs.GetReadPledge{
		FileHash:   fileHash,
		Downloader: c.DefAcc.Address,
	}
	txHash, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_CANCEL_FILE_READ, []interface{}{getReadPledge})
	if err != nil {
		return nil, err
	}

	confirmed, err := c.PollForTxConfirmed(time.Duration(common.TX_CONFIRM_TIMEOUT)*time.Second, txHash.ToArray())
	if err != nil || !confirmed {
		return txHash.ToArray(), errors.New("CancelFileRead tx is not confirmed")
	}
	return txHash.ToArray(), nil
}

func (c *Core) GenPassport(height uint32, blockHash []byte) ([]byte, error) {
	passPort := fs.Passport{
		BlockHeight: uint64(height),
		BlockHash:   blockHash,
		WalletAddr:  c.DefAcc.Address,
		PublicKey:   keypair.SerializePublicKey(c.DefAcc.PublicKey),
	}

	sinkTmp := ccom.NewZeroCopySink(nil)
	passPort.Serialization(sinkTmp)

	signData, err := common.Sign(c.DefAcc, sinkTmp.Bytes())
	if err != nil {
		return nil, fmt.Errorf("GenPassport Sign error: %s", err.Error())
	}
	passPort.Signature = signData

	sink := ccom.NewZeroCopySink(nil)
	passPort.Serialization(sink)

	return sink.Bytes(), nil
}

func (c *Core) GenFileReadSettleSlice(fileHash []byte, payTo ccom.Address, sliceId uint64,
	pledgeHeight uint64) (*fs.FileReadSettleSlice, error) {
	settleSlice := fs.FileReadSettleSlice{
		FileHash:     fileHash,
		PayFrom:      c.DefAcc.Address,
		PayTo:        payTo,
		SliceId:      sliceId,
		PledgeHeight: pledgeHeight,
	}
	sink := ccom.NewZeroCopySink(nil)
	settleSlice.Serialization(sink)

	signData, err := common.Sign(c.DefAcc, sink.Bytes())
	if err != nil {
		return nil, fmt.Errorf("FileReadSettleSlice Sign error: %s", err.Error())
	}
	settleSlice.Sig = signData
	settleSlice.PubKey = keypair.SerializePublicKey(c.DefAcc.PublicKey)
	return &settleSlice, nil
}

func (c *Core) PollForTxConfirmed(timeout time.Duration, txHash []byte) (bool, error) {
	if len(txHash) == 0 {
		return false, fmt.Errorf("txHash is empty")
	}
	txHashStr := hex.EncodeToString(ccom.ToArrayReverse(txHash))
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
