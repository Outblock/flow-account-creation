package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onflow/cadence"
	c_json "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

type Wallet struct {
	Address string `json:"address"`
	Key     string `json:"key"`
}

type Identifier [32]byte

func CreateFlowKey(hashAlgo string, signAlgo string, publicKey string, weight int, network string, c *gin.Context) *flow.Transaction {

	node := Flow(network)

	serviceAddressHex, servicePrivKeyHex, keyIndex := GetKey(network, c)
	serviceSigAlgoHex := "ECDSA_P256"

	// fmt.Printf("Service private key %s\n", servicePrivKeyHex)

	gasLimit := uint64(100)

	tx := CreateAccount(node, publicKey, signAlgo, hashAlgo, serviceAddressHex, servicePrivKeyHex, serviceSigAlgoHex, gasLimit, keyIndex, weight, network)

	return tx
}

func CreateAccount(node string,
	publicKeyHex string,
	sigAlgoName string,
	hashAlgoName string,
	serviceAddressHex string,
	servicePrivKeyHex string,
	serviceSigAlgoName string,
	gasLimit uint64,
	keyIndex int64,
	weight int,
	network string) *flow.Transaction {

	ctx := context.Background()
	log.Printf("network is: %v\n", network)

	sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoName)
	publicKey, err := crypto.DecodePublicKeyHex(sigAlgo, publicKeyHex)
	if err != nil {
		log.Println("failed to decode public key hex: ", err)
	}
	// if err != nil {
	// 	log.Println("Decode Public Key Hex Error: ", err)
	// }

	hashAlgo := crypto.StringToHashAlgorithm(hashAlgoName)

	accountKey := flow.NewAccountKey().
		SetPublicKey(publicKey).
		SetSigAlgo(sigAlgo).
		SetHashAlgo(hashAlgo).
		SetWeight(weight)

	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		log.Println("failed to connect to node")
	}

	serviceSigAlgo := crypto.StringToSignatureAlgorithm(serviceSigAlgoName)
	servicePrivKey, err := crypto.DecodePrivateKeyHex(serviceSigAlgo, servicePrivKeyHex)
	if err != nil {
		log.Println(err)
	}

	serviceAddress := flow.HexToAddress(serviceAddressHex)
	//serviceAddress := flow.HexToAddress(serviceAddressHex)
	serviceAccount, err := c.GetAccountAtLatestBlock(ctx, serviceAddress)
	if err != nil {
		log.Println(err)
	}

	serviceAccountKey := serviceAccount.Keys[keyIndex]
	serviceSigner, err := crypto.NewInMemorySigner(servicePrivKey, serviceAccountKey.HashAlgo)
	if err != nil {
		log.Println("Service Sign Error: ", err)

	}

	var tx *flow.Transaction
	if network == "previewnet" {
		tx, err = CreatePreviewnetAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)
	} else if network == "mainnet" {
		tx, err = CreateFlowMainnetAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)
	} else {
		tx, err = CreateFlowTestnetAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)

		// tx, err = templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)
	}

	// tx, err := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)
	var transactionScript string
	if network == "testnet" {
		transactionScript = `
 
 
 import Crypto
 import FungibleToken from 0x9a0766d93b6608b7
 import FlowToken from 0x7e60df042a9c0868
 
 transaction(publicKeys: [Crypto.KeyListEntry], contracts: {String: String}, fundAmount: UFix64) {
				let sentVault: @{FungibleToken.Vault}
				let signer: auth(BorrowValue | Storage) &Account
		
				prepare(signer: auth(BorrowValue, Storage) &Account) {
		
						let vaultRef = signer.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(from: /storage/flowTokenVault) ?? panic("Could not borrow reference to the owner''s Vault!")
						self.sentVault <- vaultRef.withdraw(amount: fundAmount)
		
						self.signer = signer
				}
				execute {
						let account = Account(payer: self.signer)
						for key in publicKeys {
								account.keys.add(publicKey: key.publicKey, hashAlgorithm: key.hashAlgorithm, weight: key.weight)
						}
		
						for contract in contracts.keys {
								account.contracts.add(name: contract, code: contracts[contract]!.decodeHex())
						}
						let tokenReceiver = account.capabilities.borrow<&{FungibleToken.Receiver}>(/public/flowTokenReceiver) ?? panic("Unable to borrow receiver reference")
		
						tokenReceiver.deposit(from: <-self.sentVault)
				}
		 }
	 `
	} else if network == "previewnet" {
		transactionScript = `import Crypto
		import FlowToken from 0x4445e7ad11568276
		import FungibleToken from 0xa0225e7000ac82a9
		
		 transaction(publicKeys: [Crypto.KeyListEntry], contracts: {String: String}, fundAmount: UFix64) {
				let sentVault: @{FungibleToken.Vault}
				let signer: auth(BorrowValue | Storage) &Account
		
				prepare(signer: auth(BorrowValue, Storage) &Account) {
		
						let vaultRef = signer.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(from: /storage/flowTokenVault) ?? panic("Could not borrow reference to the owner''s Vault!")
						self.sentVault <- vaultRef.withdraw(amount: fundAmount)
		
						self.signer = signer
				}
				execute {
						let account = Account(payer: self.signer)
						for key in publicKeys {
								account.keys.add(publicKey: key.publicKey, hashAlgorithm: key.hashAlgorithm, weight: key.weight)
						}
		
						for contract in contracts.keys {
								account.contracts.add(name: contract, code: contracts[contract]!.decodeHex())
						}
						let tokenReceiver = account.capabilities.borrow<&{FungibleToken.Receiver}>(/public/flowTokenReceiver) ?? panic("Unable to borrow receiver reference")
		
						tokenReceiver.deposit(from: <-self.sentVault)
				}
		 }`
	} else {
		transactionScript = `
		import Crypto
	import FlowToken from 0x1654653399040a61
  import FungibleToken from 0xf233dcee88fe0abe
	
	 transaction(publicKeys: [Crypto.KeyListEntry], contracts: {String: String}, fundAmount: UFix64) {
				let sentVault: @{FungibleToken.Vault}
				let signer: auth(BorrowValue | Storage) &Account
		
				prepare(signer: auth(BorrowValue, Storage) &Account) {
		
						let vaultRef = signer.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(from: /storage/flowTokenVault) ?? panic("Could not borrow reference to the owner''s Vault!")
						self.sentVault <- vaultRef.withdraw(amount: fundAmount)
		
						self.signer = signer
				}
				execute {
						let account = Account(payer: self.signer)
						for key in publicKeys {
								account.keys.add(publicKey: key.publicKey, hashAlgorithm: key.hashAlgorithm, weight: key.weight)
						}
		
						for contract in contracts.keys {
								account.contracts.add(name: contract, code: contracts[contract]!.decodeHex())
						}
						let tokenReceiver = account.capabilities.borrow<&{FungibleToken.Receiver}>(/public/flowTokenReceiver) ?? panic("Unable to borrow receiver reference")
		
						tokenReceiver.deposit(from: <-self.sentVault)
				}
		 }`
	}

	s := map[string]interface{}{}
	s["type"] = "UFix64"
	s["value"] = "0.00100000"

	cv, err := ArgAsCadence(s)
	if err != nil {
		log.Println(err)
	}

	err = tx.AddArgument(cv)
	if err != nil {
		log.Println(err)
	}

	tx.SetScript([]byte(transactionScript))

	if err != nil {
		log.Println("Create Account Error: ", err)
	}
	tx.SetProposalKey(serviceAddress, serviceAccountKey.Index, serviceAccountKey.SequenceNumber)
	tx.SetPayer(serviceAddress)
	tx.SetGasLimit(uint64(gasLimit))

	latestBlock, err := c.GetLatestBlockHeader(context.Background(), true)
	tx.SetReferenceBlockID(latestBlock.ID)

	err = tx.SignEnvelope(serviceAddress, serviceAccountKey.Index, serviceSigner)
	err = c.SendTransaction(context.Background(), *tx)
	if err != nil {
		log.Println(err)
	}
	return tx
}

func WaitForSeal(id flow.Identifier, network string) string {
	ctx := context.Background()

	node := Flow(network)
	c, connectErr := client.New(node, grpc.WithInsecure())
	if connectErr != nil {
		log.Println("failed to connect to node")
	}

	result, err := c.GetTransactionResult(ctx, id)
	Handle(err)

	fmt.Printf("Waiting for transaction %s to be sealed...\n", id)

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		fmt.Print(".")
		result, err = c.GetTransactionResult(ctx, id)
		Handle(err)
	}

	fmt.Printf("Transaction %s sealed\n", id)
	newAddress := findAccount(ctx, c, id)
	addressString := flow.Address.String(newAddress)
	return addressString
}

func findAccount(ctx context.Context, c *client.Client, id flow.Identifier) flow.Address {

	result, err := c.GetTransactionResult(ctx, id)
	if err != nil {
		log.Println("failed to get transaction result")
	}

	var newAddress flow.Address

	if result.Status != flow.TransactionStatusSealed {
		log.Println("address not known until transaction is sealed")
	}

	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			newAddress = flow.AccountCreatedEvent(event).Address()
			break
		}
	}

	fmt.Println()
	fmt.Printf("New address -> %s", newAddress)
	return newAddress
}

func Handle(err error) {

	if err != nil {
		// fmt.Println("err:", err.Error())
		log.Println(err)
	}
}

func DecodePublicKey(publicKey string) (crypto.PublicKey, error) {
	pk, error := crypto.DecodePublicKeyHex(crypto.ECDSA_secp256k1, publicKey)

	return pk, error
}

func SendTransaction() {
	flow.NewTransaction()
}

type noopMemoryGauge struct{}

func (g noopMemoryGauge) Use(_ common.MemoryUsage) {}

func (g noopMemoryGauge) UseMemory(_ uint64) {}

func (g noopMemoryGauge) MeterMemory(_ common.MemoryUsage) error {
	return nil
}

func ArgAsCadence(a interface{}) (cadence.Value, error) {
	c, ok := a.(cadence.Value)
	if ok {
		return c, nil
	}
	// Convert to json bytes so we can use cadence's own encoding library
	j, err := json.Marshal(a)
	if err != nil {
		return cadence.Void{}, err
	}
	gauge := noopMemoryGauge{}
	// Use cadence's own encoding library
	c, err = c_json.Decode(gauge, j)

	if err != nil {
		return cadence.Void{}, err
	}
	return c, nil
}
