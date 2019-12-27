package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	apiComm "github.com/ontio/ontfs-contract-api/common"
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

type Core struct {
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

func Init(walletPath string, walletPwd string, ontRpcSrvAddr string, gasPrice uint64, gasLimit uint64) *Core {
	contractAddr = utils.OntFSContractAddress
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
		src := common.NewZeroCopySource(retInfo.Info)
		if err = globalParam.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetGlobalParam error: %s", err.Error())
		}
		return &globalParam, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) NodeRegister(volume uint64, serviceTime uint64, nodeNetAddr string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeRegister DefAcc is nil")
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_REGISTER, []interface{}{&fs.FsNodeInfo{Pledge: 0, Profit: 0, Volume: volume, RestVol: 0,
			ServiceTime: serviceTime, NodeAddr: c.WalletAddr, NodeNetAddr: []byte(nodeNetAddr)}})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) NodeQuery(nodeWallet common.Address) (*fs.FsNodeInfo, error) {
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
		src := common.NewZeroCopySource(retInfo.Info)
		if err = fsNodeInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("NodeQuery error: %s", err.Error())
		}
		return &fsNodeInfo, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) NodeUpdate(volume uint64, serviceTime uint64, nodeNetAddr string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeUpdate DefAcc is nil")
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_UPDATE, []interface{}{&fs.FsNodeInfo{Pledge: 0, Profit: 0, Volume: volume, RestVol: 0,
			ServiceTime: serviceTime, NodeAddr: c.WalletAddr, NodeNetAddr: []byte(nodeNetAddr)}},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) NodeCancel() ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeCancel DefAcc is nil")
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_CANCEL, []interface{}{c.WalletAddr})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) NodeWithDrawProfit() ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("NodeWithDrawProfit DefAcc is nil")
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_NODE_WITH_DRAW_PROFIT, []interface{}{c.WalletAddr},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
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
		src := common.NewZeroCopySource(retInfo.Info)
		if err = fileInfo.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFileInfo error: %s", err.Error())
		}
		return &fileInfo, err
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) FileProve(fileHashStr string, pdpVersion uint64, multiRes []byte, addResStr string,
	blockHeight uint64) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}
	fileHash := []byte(fileHashStr)
	addRes := []byte(addResStr)
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_FILE_PROVE, []interface{}{&fs.PdpData{
			Version:         pdpVersion,
			FileHash:        fileHash,
			NodeAddr:        c.WalletAddr,
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

func (c *Core) GetFileReadPledge(fileHashStr string, downloader common.Address) (*fs.ReadPledge, error) {
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

func (c *Core) FileReadProfitSettle(fileReadSettleSlice *fs.FileReadSettleSlice) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("FileReadProfitSettle DefAcc is nil")
	}
	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_READ_FILE_SETTLE, []interface{}{fileReadSettleSlice},
	)
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) VerifyFileReadSettleSlice(settleSlice *fs.FileReadSettleSlice) (bool, error) {
	tmpSettleSlice := fs.FileReadSettleSlice{
		FileHash:     settleSlice.FileHash,
		PayFrom:      settleSlice.PayFrom,
		PayTo:        settleSlice.PayTo,
		SliceId:      settleSlice.SliceId,
		PledgeHeight: settleSlice.PledgeHeight,
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
		src := common.NewZeroCopySource(retInfo.Info)
		if err = pdpRecordList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFilePdpRecordList deserialize error: %s", err.Error())
		}
		return &pdpRecordList, err
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
}

func (c *Core) GenChallenge(nodeAddr common.Address, hash []byte, fileBlockNum, proveNum uint64) []pdp.Challenge {
	return fs.GenChallenge(nodeAddr, hash, uint32(fileBlockNum), uint32(proveNum))
}

func (c *Core) GetNodeInfo(nodeWallet common.Address) (*fs.FsNodeInfo, error) {
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
		src := common.NewZeroCopySource(retInfo.Info)
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
		src := common.NewZeroCopySource(retInfo.Info)
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

	spaceInfo := fs.SpaceInfo{
		SpaceOwner:  c.DefAcc.Address,
		Volume:      volume,
		CopyNumber:  copyNumber,
		PdpInterval: pdpInterval,
		TimeExpired: timeExpired,
	}

	sink := common.NewZeroCopySink(nil)
	spaceInfo.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion,
		contractAddr, fs.FS_CREATE_SPACE, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
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
		src := common.NewZeroCopySource(retInfo.Info)
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

	sink := common.NewZeroCopySink(nil)
	spaceUpdate.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion,
		contractAddr, fs.FS_UPDATE_SPACE, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) DeleteSpace() ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_DELETE_SPACE, []interface{}{c.DefAcc.Address})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) StoreFile(fileHash string, fileBlockCount uint64, pdpInterval uint64, timeExpired uint64, copyNum uint64,
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
		PdpInterval:    pdpInterval,
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

func (c *Core) RenewFile(fileHashStr string, renewTimes uint64) ([]byte, error) {
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

func (c *Core) DeleteFiles(fileHashStrs []string) ([]byte, error) {
	if c.DefAcc == nil {
		return nil, errors.New("DefAcc is nil")
	}

	var fileDelList fs.FileDelList
	for _, fileHashStr := range fileHashStrs {
		fileDelList.FilesDel = append(fileDelList.FilesDel, fs.FileDel{FileHash: []byte(fileHashStr)})
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

func (c *Core) ChangeFileOwner(fileHashStr string, newOwner common.Address) ([]byte, error) {
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
		src := common.NewZeroCopySource(retInfo.Info)
		if err = fileList.Deserialization(src); err != nil {
			return nil, fmt.Errorf("GetFileList error: %s", err.Error())
		}
		return &fileList, nil
	} else {
		return nil, errors.New(string(retInfo.Info))
	}
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

	sink := common.NewZeroCopySink(nil)
	fileReadPledge.Serialization(sink)

	ret, err := c.OntSdk.Native.InvokeNativeContract(c.GasPrice, c.GasLimit, c.DefAcc, contractVersion, contractAddr,
		fs.FS_READ_FILE_PLEDGE, []interface{}{sink.Bytes()})
	if err != nil {
		return nil, err
	}
	return ret.ToArray(), err
}

func (c *Core) GenPassport(height uint32, blockHash []byte) ([]byte, error) {
	passPort := fs.Passport{
		BlockHeight: uint64(height),
		BlockHash:   blockHash,
		WalletAddr:  c.DefAcc.Address,
		PublicKey: keypair.SerializePublicKey(c.DefAcc.PublicKey),
	}

	sinkTmp := common.NewZeroCopySink(nil)
	passPort.Serialization(sinkTmp)

	signData, err := apiComm.Sign(c.DefAcc, sinkTmp.Bytes())
	if err != nil {
		return nil, fmt.Errorf("GenPassport Sign error: %s", err.Error())
	}
	passPort.Signature = signData

	sink := common.NewZeroCopySink(nil)
	passPort.Serialization(sink)

	return sink.Bytes(), nil
}

func (c *Core) GenFileReadSettleSlice(fileHash []byte, payTo common.Address, sliceId uint64,
	pledgeHeight uint64) (*fs.FileReadSettleSlice, error) {
	settleSlice := fs.FileReadSettleSlice{
		FileHash:     fileHash,
		PayFrom:      c.DefAcc.Address,
		PayTo:        payTo,
		SliceId:      sliceId,
		PledgeHeight: pledgeHeight,
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

func (c *Core) CancelFileRead(fileHashStr string) ([]byte, error) {
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

func (c *Core) PollForTxConfirmed(timeout time.Duration, txHash []byte) (bool, error) {
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
