package main

import (
	"github.com/ontio/ontfs-contract-api/common"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs"
	"log"
	"net"
	"strings"
	"time"
)

const TestFileHash = "FileTest"

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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 2048)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println(conn.RemoteAddr().String(), "connection error: ", err)
			return
		}

		msg := string(buffer[:n])
		log.Println(conn.RemoteAddr().String(), "receive data string:\n", msg)

		parts := strings.Split(msg, "|")
		if 0 == strings.Compare(parts[0], "StoreFile") {
			go PDP(parts[1])
		}

		conn.Write([]byte("Message Received"))
	}
}

func PDP(fileStoreTxHash string) {
	log.Printf("PDP init, FileStoreTxHash: [%s]", fileStoreTxHash)

	fileInfo, err := fsCore.GetFileInfo(TestFileHash)
	if err != nil {
		log.Printf("GetFileInfo error: %s", err.Error())
		return
	}
	filePdpNeedCount := (fileInfo.TimeExpired-fileInfo.TimeStart)/ontfs.DefaultPdpInterval + 1
	log.Printf("TotalPdpNeedCount: %d", filePdpNeedCount)
	common.PrintStruct(*fileInfo)

	fileHashStr := string(fileInfo.FileHash)

	log.Printf("FileProve first time")
	_, err = fsCore.FileProve(fileHashStr, []byte("test"), "test", 8)
	if err != nil {
		log.Printf("First FileProve error: %s", err.Error())
	}

	for {
		time.Sleep(ontfs.DefaultPdpInterval * time.Second)
		fileInfo1, err := fsCore.GetFileInfo(TestFileHash)
		if err != nil && fileInfo1 != nil{
			log.Printf("File is not exist. Return")
			return
		}

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
				_, err = fsCore.FileProve(fileHashStr, []byte(TestFileHash), TestFileHash, pdpInfo.NextHeight)
				if err != nil {
					log.Printf("FileProve error: %s", err.Error())
				}
				log.Printf("FileProve end")
			}
		}
	}
}
