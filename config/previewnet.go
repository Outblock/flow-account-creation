/*
 * Flow Go SDK
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"encoding/hex"
	"fmt"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
)

type Contract struct {
	Name   string
	Source string
}

func (c Contract) SourceBytes() []byte {
	return []byte(c.Source)
}

func (c Contract) SourceHex() string {
	return hex.EncodeToString(c.SourceBytes())
}

var previewAccountCreate = []byte(`
 import Crypto
 
 transaction(publicKeys: [Crypto.KeyListEntry], contracts: {String: String}) {
		 prepare(signer: auth(BorrowValue | Storage) &Account) {
 
				 let account = Account(payer: signer)
 
				 // add all the keys to the account
				 for key in publicKeys {
						 account.keys.add(publicKey: key.publicKey, hashAlgorithm: key.hashAlgorithm, weight: key.weight)
				 }
 
				 // add contracts if provided
				 for contract in contracts.keys {
						 account.contracts.add(name: contract, code: contracts[contract]!.decodeHex())
				 }
		 }
 }
	`)

func CreatePreviewnetAccount(accountKeys []*flow.AccountKey, contracts []Contract, payer flow.Address) (*flow.Transaction, error) {
	keyList := make([]cadence.Value, len(accountKeys))

	contractKeyPairs := make([]cadence.KeyValuePair, len(contracts))

	var err error
	for i, key := range accountKeys {
		keyList[i], err = templates.AccountKeyToCadenceCryptoKey(key)
		if err != nil {
			return nil, fmt.Errorf("cannot create CreateAccount transaction: %w", err)
		}
	}

	for i, contract := range contracts {
		contractKeyPairs[i] = cadence.KeyValuePair{
			Key:   cadence.String(contract.Name),
			Value: cadence.String(contract.SourceHex()),
		}
	}

	cadencePublicKeys := cadence.NewArray(keyList)
	cadenceContracts := cadence.NewDictionary(contractKeyPairs)

	return flow.NewTransaction().
		SetScript(previewAccountCreate).
		AddAuthorizer(payer).
		AddRawArgument(jsoncdc.MustEncode(cadencePublicKeys)).
		AddRawArgument(jsoncdc.MustEncode(cadenceContracts)), nil
}
