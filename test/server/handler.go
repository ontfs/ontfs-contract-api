package main

import (
	"encoding/hex"
	"github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontology-crypto/pdp"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"log"
	"net"
	"strings"
	"time"
)

func FsServer() {
	netListen, err := net.Listen("tcp", "localhost:1024")
	if err != nil {
		log.Println("net Listen error: ", err.Error())
		return
	}

	defer netListen.Close()
	log.Println("Waiting for clients ...")

	for {
		conn, err := netListen.Accept()
		if err != nil {
			continue
		}

		log.Println(conn.RemoteAddr().String(), "tcp connect success")
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 2048)

	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(conn.RemoteAddr().String(), "Connection error: ", err)
		return
	}
	msg := string(buffer[:n])
	log.Println(conn.RemoteAddr().String(), "Receive data:\n", msg)
	conn.Write([]byte("Message Received"))

	parts := strings.Split(msg, "|")
	if 0 == strings.Compare(parts[0], "StoreFile") {
		go PDP(parts[1])
	} else if 0 == strings.Compare(parts[0], "ReadFile") {
		go FileRead(conn, parts[1], parts[2])
	}
}

func PDP(fileHash string) {
	log.Printf("PDP init, FileHash: [%s]", fileHash)

	fileInfo, err := fsCore.GetFileInfo(fileHash)
	if err != nil {
		log.Printf("GetFileInfo error: %s", err.Error())
		return
	}
	filePdpNeedCount := (fileInfo.TimeExpired-fileInfo.TimeStart)/fileInfo.PdpInterval + 1
	log.Printf("TotalPdpNeedCount: %d", filePdpNeedCount)
	common.PrintStruct(*fileInfo)

	fileHashStr := string(fileInfo.FileHash)

	log.Printf("FileProve first time")
	_, err = fsCore.FileProve(fileHashStr, pdp.Version, []byte("test"), "test", 8)
	if err != nil {
		log.Printf("First FileProve error: %s", err.Error())
	}

	for {
		time.Sleep(time.Duration(fileInfo.PdpInterval * uint64(time.Second)))
		fileInfo1, err := fsCore.GetFileInfo(fileHash)
		if err != nil || fileInfo == nil {
			log.Printf("File is not exist. Return")
			return
		}
		filePdpNeedCount := (fileInfo1.TimeExpired-fileInfo1.TimeStart)/fileInfo.PdpInterval + 1
		log.Printf("TotalPdpNeedCount: %d", filePdpNeedCount)

		log.Printf("FileProve begin")
		pdpInfoList, err := fsCore.GetFilePdpRecordList(fileHashStr)
		if err != nil {
			log.Printf("GetFilePdpRecordList error: %s", err.Error())
			break
		}
		if len(pdpInfoList.PdpRecords) == 0 {
			log.Println("GetFilePdpRecordList error: PdpRecords Length is 0")
			break
		}

		for _, pdpInfo := range pdpInfoList.PdpRecords {
			if pdpInfo.NodeAddr == fsCore.WalletAddr {
				common.PrintStruct(pdpInfo)
				_, err = fsCore.FileProve(fileHashStr, pdp.Version, []byte(fileHash), fileHash, pdpInfo.NextHeight)
				if err != nil {
					log.Printf("FileProve error: %s", err.Error())
				}
				log.Printf("FileProve end")
			}
		}
	}
}

func FileRead(conn net.Conn, fileHash string, downloader string) {
	buffer := make([]byte, 2048)
	var fileReadSettleSlice *ontfs.FileReadSettleSlice
	downloaderAddr, err := utils.AddressFromBase58(downloader)
	if err != nil {
		log.Printf("FileRead AddressFromBase58 error: %s", err.Error())
		return
	}

	readPledge, err := fsCore.GetFileReadPledge(fileHash, downloaderAddr)
	if err != nil {
		log.Printf("FileRead GetFileReadPledge error: %s", err.Error())
		return
	}
	common.PrintStruct(*readPledge)

	for _, readPlan := range readPledge.ReadPlans {
		if readPledge.ReadPlans[0].NodeAddr.ToBase58() == fsCore.WalletAddr.ToBase58() {
			for i := uint64(0); i < readPlan.MaxReadBlockNum; i++ {
				if i+readPlan.HaveReadBlockNum >= readPlan.MaxReadBlockNum {
					log.Println("FileReadPledge is not valid")
					return
				}

				n, err := conn.Read(buffer)
				if err != nil {
					log.Println(conn.RemoteAddr().String(), "Connection error: ", err.Error())
					return
				}
				log.Println("Received FileReadSettleSlice")
				msg := string(buffer[:n])
				sliceData, err := hex.DecodeString(msg)
				if err != nil {
					log.Println("DecodeString error: ", err)
					return
				}

				fileReadSettleSlice, err = common.FileReadSettleSliceDeserialize(sliceData)
				if err != nil {
					log.Println("FileReadSettleSliceDeserialize error: ", err.Error())
					return
				}
				ret, err := fsCore.VerifyFileReadSettleSlice(fileReadSettleSlice)
				if err != nil {
					log.Println("VerifyFileReadSettleSlice error: ", err.Error())
					return
				}
				if !ret {
					log.Println("VerifyFileReadSettleSlice failed")
					return
				}
				conn.Write([]byte("FileReadSettleSlice Received"))
				log.Println("Send FileReadSettleSlice ACK")
			}
			log.Println("FileReadProfitSettle...")
			settleTx, err := fsCore.FileReadProfitSettle(fileReadSettleSlice)
			if err != nil {
				log.Println("FileReadProfitSettle failed")
				return
			}
			fsCore.PollForTxConfirmed(14*time.Second, settleTx)
			log.Println("FileReadProfitSettle over")
		}
	}
}
