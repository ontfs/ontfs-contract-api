package common

import (
	"errors"
	"fmt"
	"reflect"

	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	ont "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"strings"
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
		return nil, fmt.Errorf("PdpParamDeserialize error: %s", err.Error())
	}
	return &pdpParam, nil
}

func FileReadSettleSliceSerialize(fileReadSettleSlice *fs.FileReadSettleSlice) []byte {
	sink := common.NewZeroCopySink(nil)
	fileReadSettleSlice.Serialization(sink)
	return sink.Bytes()
}

func FileReadSettleSliceDeserialize(fileReadSettleSliceData []byte) (*fs.FileReadSettleSlice, error) {
	var fileReadSettleSlice fs.FileReadSettleSlice
	src := common.NewZeroCopySource(fileReadSettleSliceData)
	if err := fileReadSettleSlice.Deserialization(src); err != nil {
		return nil, fmt.Errorf("FileReadSettleSliceDeserialize error: %s", err.Error())
	}
	return &fileReadSettleSlice, nil
}

func PrintStruct(st interface{}) {
	dataType := reflect.TypeOf(st)
	dataValue := reflect.ValueOf(st)
	fmt.Printf("[=====%s======]\n", dataType.Name())

	num := dataType.NumField()
	for id := 0; id < num; id++ {
		field := dataType.Field(id)
		fieldName := field.Name
		value := dataValue.FieldByName(fieldName)
		if 0 == strings.Compare(fieldName, "NodeNetAddr") {
			fmt.Printf("-[%-20s]:\t %s\n", fieldName, value)
		} else if 0 == strings.Compare(fieldName, "NodeAddr") ||
			0 == strings.Compare(fieldName, "FileOwner") ||
			0 == strings.Compare(fieldName, "SpaceOwner")      {
			hexAddr := fmt.Sprintf("%x", value)
			addr, err := hex.DecodeString(hexAddr)
			if err != nil {
				fmt.Printf("PrintStruct error NodeAddr DecodeString error")
				continue
			} else {
				var nodeAddr common.Address
				copy(nodeAddr[:], addr[:])
				fmt.Printf("-[%-20s]:\t %s\n", fieldName, nodeAddr.ToBase58())
			}
		} else if reflect.TypeOf(value).String() == "struct" {
			PrintStruct(value)
		} else {
			fmt.Printf("-[%-20s]:\t %v\n", fieldName, value)
		}
	}
}
