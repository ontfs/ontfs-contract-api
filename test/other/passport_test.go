package other

import (
	"testing"

	fs "github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"github.com/ontio/ontfs-contract-api/client"
)

func TestPassport_Check(t *testing.T) {
	client := client.Init("./wallet.dat", "pwd", "")
	if client == nil {
		t.Fatalf("client init error")
	}

	blockHash := []byte{0x01, 0x02, 0x03}
	passport, err := client.GenPassport(1, blockHash)
	if err != nil {
		t.Fatalf("client GenPassport error")
	}

	walletAddr, err := fs.CheckPassport(5, passport)
	if err != nil {
		t.Fatalf("client CheckPassport error: %s", err.Error())
	}
	t.Logf("%s", walletAddr.ToBase58())
}
