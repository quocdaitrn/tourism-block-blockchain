package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	// "time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("tourism_block")

	// log.Println("--> Submit Transaction: CreateService")
	// result, err := contract.SubmitTransaction("CreateService", "5f793bd99b0afd906562d391")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Submit Transaction: ReadService")
	// result, err := contract.SubmitTransaction("ReadService", "5f793bd99b0afd906562d390")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Submit Transaction: AddAgreement")
	// result, err := contract.SubmitTransaction("AddAgreement", "5f793bd99b0afd906562d390", "5f82dbd1ae82399ce1d95f61", "true")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Submit Transaction: UpdateAgreement")
	// result, err := contract.SubmitTransaction("UpdateAgreement", "5f793bd99b0afd906562d390", "5f82dbd1ae82399ce1d95f60", "false")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// log.Println("--> Submit Transaction: RemoveAgreement")
	// result, err := contract.SubmitTransaction("RemoveAgreement", "5f793bd99b0afd906562d390", "5f82dbd1ae82399ce1d95f61")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// now := time.Now()
	// at := now.Format("2006-01-02T15:04:05.000Z")
	// log.Println("--> Submit Transaction: HandleSatisfactionEvaluationEvent")
	// result, err := contract.SubmitTransaction("HandleSatisfactionEvaluationEvent", "5f793bd99b0afd906562d390", "5f82dbd1ae82399ce1d95f60", "5f82e354a2100c7d7dc1a190", "6d976a7b3ef96ed1334de159eeb4aaac", at, "false")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	// now := time.Now()
	// at := now.Format("2006-01-02T15:04:05.000Z")
	// log.Println("--> Submit Transaction: HandlePenaltyRuleEvaluationEvent")
	// result, err := contract.SubmitTransaction("HandlePenaltyRuleEvaluationEvent", "5f793bd99b0afd906562d390", "5f82dbd1ae82399ce1d95f60", "5f82e354a2100c7d7dc1a191", "6d976a7b3ef96ed1334de159eeb4aaad", at, "true")
	// if err != nil {
	// 	log.Fatalf("Failed to submit transaction: %v", err)
	// }
	// log.Println(string(result))

	log.Println("--> Evaluate Transaction: CountAllEvaluations")
	result, err := contract.EvaluateTransaction("CountAllEvaluations", "1000")
	if err != nil {
		log.Fatalf("failed to submit transaction: %v\n", err)
	}
	log.Println(string(result))
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"/",
		"Users",
		"daitran",
		"Projects",
		"master",
		"tourism-block-blockchain",
		"network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
