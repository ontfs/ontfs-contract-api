package other

import (
	"fmt"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"
)

func TestPdpParamSer(t *testing.T) {
	pdpParam := &ontfs.PdpParam{
		Version: 1,
		G:       []byte{0x01},
		G0:      []byte{0x02},
		PubKey:  []byte{0x03},
		FileId:  []byte{0x04},
	}
	sink := common.NewZeroCopySink(nil)
	pdpParam.Serialization(sink)

	fmt.Printf("%v\n", sink.Bytes())
	fmt.Printf("%d\n", sink.Size())
	buf := make([]byte, sink.Size())
	copy(buf, sink.Bytes())
	fmt.Printf("%v\n", buf)

	var pdpTmp ontfs.PdpParam

	source := common.NewZeroCopySource(buf)
	err := pdpTmp.Deserialization(source)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Printf("%v\n", pdpTmp)
}
