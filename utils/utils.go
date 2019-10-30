package utils

import (
	"errors"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	ont "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
)

func Sign(acc *ont.Account, data []byte) ([]byte, error) {
	sig, err := s.Sign(s.SignatureScheme(acc.SigScheme), acc.PrivateKey, data, nil)
	if err != nil {
		return nil, err
	}
	sigData, err := s.Serialize(sig)
	if err != nil {
		return nil, fmt.Errorf("signature.Serialize error:%s", err)
	}
	return sigData, nil
}

// Verify check the signature of data using pubKey
func Verify(pubKey keypair.PublicKey, data, signature []byte) error {
	sigObj, err := s.Deserialize(signature)
	if err != nil {
		return errors.New("invalid signature data: " + err.Error())
	}
	if !s.Verify(pubKey, data, sigObj) {
		return errors.New("signature verification failed")
	}
	return nil
}

func PdpParamSerialize(g []byte, g0 []byte, pubKey []byte, fileId []byte) []byte {
	pdpParam := fs.PdpParam{
		G:      g,
		G0:     g0,
		PubKey: pubKey,
		FileId: fileId,
	}

	sink := common.NewZeroCopySink(nil)
	pdpParam.Serialization(sink)
	return sink.Bytes()
}

func PdpParamDeserialize(pdpParamData []byte) (*fs.PdpParam, error) {
	var pdpParam fs.PdpParam
	src := common.NewZeroCopySource(pdpParamData)
	if err := pdpParam.Deserialization(src); err != nil {
		return nil, fmt.Errorf("PdpParamDeserialize Deserialize error: %s", err.Error())
	}
	return &pdpParam, nil
}

func FileReadSettleSliceSerialize(fileHash []byte, payFrom common.Address,
	payTo common.Address, sliceId uint64, sig []byte, pubKey []byte) []byte {
	fileReadSettleSlice := fs.FileReadSettleSlice{
		FileHash: fileHash,
		PayFrom:  payFrom,
		PayTo:    payTo,
		SliceId:  sliceId,
		Sig:      sig,
		PubKey:   pubKey,
	}
	sink := common.NewZeroCopySink(nil)
	fileReadSettleSlice.Serialization(sink)
	return sink.Bytes()
}

func FileReadSettleSliceDeserialize(fileReadSettleSliceData []byte) (*fs.FileReadSettleSlice, error) {
	var fileReadSettleSlice fs.FileReadSettleSlice
	src := common.NewZeroCopySource(fileReadSettleSliceData)
	if err := fileReadSettleSlice.Deserialization(src); err != nil {
		return nil, fmt.Errorf("FileReadSettleSlice Deserialize error: %s", err.Error())
	}
	return &fileReadSettleSlice, nil
}
