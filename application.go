package main

import (
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	var _LOGGER = log.New(os.Stderr, "ApplicationLogger ", log.Lshortfile)
	rand.Seed(time.Now().UnixNano())
	const letters = "0123456789"
	_LOGGER.Println("Starting the testing of the SDKClient")
	hlfClient := HLFApplicationClient{
		ConfigFile: "config_org1.yaml",
		ChannelID:  "mychannel1",
		OrgID:      "org1",
	}
	if err := hlfClient.Initialize(); err != nil {
		_LOGGER.Printf("Initialization of client failed %v\n", err)
		return
	}
	_, err := hlfClient.GetChannel("mychannel1", []string{"org1"})
	if err != nil {
		_LOGGER.Printf("Channel %s setup error  %v\n", "mychannel1", err)
		return
	}
	_LOGGER.Println("Channel setup done")
	randomNumber := make([]byte, 4)
	for i := range randomNumber {
		randomNumber[i] = letters[rand.Intn(len(letters))]
	}
	_LOGGER.Printf("Going to store %s", string(randomNumber))
	var txArgs = [][]byte{[]byte("X"), randomNumber}
	trxnIdInvoke, _ := hlfClient.InvokeTransaction("first", "put", txArgs, []string{"peer0.org1.example.com"})
	time.Sleep(time.Second * 10)
	var qryArgs = [][]byte{[]byte("X")}
	hlfClient.InvokeQueryTranaction("first", "get", qryArgs, []string{"peer0.org1.example.com"})
	hlfClient.GenerateStatistics("peer0.org1.example.com", []string{trxnIdInvoke})
}
