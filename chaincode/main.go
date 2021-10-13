package main

import (
	"log"

	"bitbucket.org/quocdaitrn/tourism-block-blockchain/chaincode/smartcontract"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	tourismChaincode, err := contractapi.NewChaincode(&smartcontract.SmartContract{})
	if err != nil {
		log.Panicf("Error creating Tourism chaincode: %v", err)
	}

	if err := tourismChaincode.Start(); err != nil {
		log.Panicf("Error starting Tourism chaincode: %v", err)
	}
}
