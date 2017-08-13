/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Transaction with metadata, always stored in accounts
type Transaction struct {
	SessionID        string
	Cpo              string
	Emp              string
	Product          string
	EvseID           string
	UserID           string
	Timestamp        string
	ChargingDuration float32
	ChargedEnergy    float32
	PricePerUnit     float32
	ValueBrutto      int
}

// Account - every transaction belongs to the receiving and the transmitting account
// an acount contains the total account balance and the history of transactions affecting the account
type Account struct {
	// balance of the account including taxation etc.
	BalanceBrutto int
	//array of transactions an account had
	Transactions []Transaction
}

// Init - called once when deploying chaincode
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Maraike said: Init called, initializing chaincode\n")

	// entity keys / identifiers
	var aKey, bKey string
	var aVal, bVal int

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	var err error

	// set involved EMP and CPO from function call
	aKey = args[0]
	aVal, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, err
	}
	aAccount := Account{}
	aAccount.BalanceBrutto = aVal
	aAccountBytes, _ := json.Marshal(aAccount)

	// Write the state to the ledger
	err = stub.PutState(aKey, aAccountBytes)
	if err != nil {
		return nil, err
	}

	fmt.Printf("account %s is persisted with balance %d\n", aKey, aAccount.BalanceBrutto)

	bKey = args[2]
	bVal, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, err
	}
	bAccount := Account{}
	bAccount.BalanceBrutto = bVal
	bAccountBytes, _ := json.Marshal(bAccount)

	// Write the state to the ledger
	err = stub.PutState(bKey, bAccountBytes)
	if err != nil {
		return nil, err
	}

	fmt.Printf("account %s is persisted with balance %d\n", bKey, bAccount.BalanceBrutto)

	// read accounts just persisted above from blockchain
	account1 := Account{}
	aAccountBytes2, _ := stub.GetState(aKey)
	json.Unmarshal(aAccountBytes2, &account1)
	fmt.Printf("account %s is read with balance %d\n", aKey, account1.BalanceBrutto)

	account2 := Account{}
	bAccountBytes2, _ := stub.GetState(bKey)
	json.Unmarshal(bAccountBytes2, &account2)
	fmt.Printf("account %s is read with balance %d\n", bKey, account2.BalanceBrutto)

	return nil, nil
}

// ============================================================================================================================
// Transaction: payment of X euro cents from EMP to CPO
// ============================================================================================================================
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	fmt.Printf("Running invoke\n")

	var err error

	if len(args) != 1 {
		return nil, errors.New("Invoke: Expecting one argument of type Transaction")
	}

	// unmarshall transaction from JSON format to Transaction object
	transaction := Transaction{}
	err = json.Unmarshal([]byte(args[0]), &transaction)
	if err != nil {
		return nil, errors.New("Invoke: Cannot unmarshal " + args[0])
	}

	// empKey and cpoKey are the account names stored in the blockchain
	empKey := transaction.Emp
	cpoKey := transaction.Cpo

	// the monetary value of the transaction
	transactionValue := transaction.ValueBrutto

	// EMP and CPO objects of type Account loaded from blockchain
	var empAccount, cpoAccount Account
	// updated accounts later stored back to the blockchain
	var empAccountBytes, cpoAccountBytes []byte

	// load EMPs account (or create a new one, in case this is the first transaction involving this EMP)
	empAccount, err = t.getOrCreateNewAccount(stub, empKey)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Account balance of %s prior transaction is %d Eurocents\n", empKey, empAccount.BalanceBrutto)

	// load CPOs account (or create a new one, in case this is the first transaction involving this CPO)
	cpoAccount, err = t.getOrCreateNewAccount(stub, cpoKey)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Account balance of %s prior transaction is %d Eurocents\n", cpoKey, cpoAccount.BalanceBrutto)

	// calculate the new total account balances for EMP and CPO
	empAccount.BalanceBrutto = empAccount.BalanceBrutto - transactionValue
	cpoAccount.BalanceBrutto = cpoAccount.BalanceBrutto + transactionValue

	// add current transaction to the EMPs and the CPOs transaction list
	empAccount.Transactions = append(empAccount.Transactions, transaction)
	cpoAccount.Transactions = append(cpoAccount.Transactions, transaction)

	//fmt.Printf("EMP_balance = %d, CPO_balance = %d\n", EMP_balance, CPO_balance)
	fmt.Printf("Account balance of %s after transaction is %d Eurocents; the account now contains %d transactions\n", empKey, empAccount.BalanceBrutto, len(empAccount.Transactions))
	fmt.Printf("Account balance of %s after transaction is %d Eurocents; the account now contains %d transactions\n", cpoKey, cpoAccount.BalanceBrutto, len(empAccount.Transactions))

	// write the updated EMP account back to the ledger
	empAccountBytes, err = json.Marshal(empAccount)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(empKey, empAccountBytes)
	if err != nil {
		return nil, err
	}

	// write the updated CPO acccount back to the ledger
	cpoAccountBytes, err = json.Marshal(cpoAccount)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(cpoKey, cpoAccountBytes)
	if err != nil {
		return nil, err
	}

	fmt.Printf("completed invoke successfully\n")

	// function ran through without errors
	return nil, nil
}

// ============================================================================================================================
// read an account from the blockchain; if it doesn't exist yet, create a new one beforehand; return the account
// ============================================================================================================================
func (t *SimpleChaincode) getOrCreateNewAccount(stub shim.ChaincodeStubInterface, accountKey string) (Account, error) {
	var jsonResp string
	// account object as stored in blockchain
	var accountValueBytes []byte
	var err error
	// empty account template
	var account Account
	account = Account{}

	accountValueBytes, err = stub.GetState(accountKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + accountKey + "\"}"
		return account, errors.New(jsonResp)
	}

	// if loading account returned no bytes, no account exists
	// --> create an account and load new accounts object bytes
	if accountValueBytes == nil {
		fmt.Printf("%s has no account in the ledger yet\n", accountKey)
		account.BalanceBrutto = 0
		accountValueBytes, err = json.Marshal(account)
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to marshal json for new account " + accountKey + "\"}"
			return account, errors.New(jsonResp)
		}
	}

	// fill account template with values read from blockchain
	json.Unmarshal(accountValueBytes, &account)

	// return account object with values
	return account, nil
}

// ============================================================================================================================
// Deletes an entity from state
// ============================================================================================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running delete\n")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	accountKey := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(accountKey)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}

// Invoke callback representing the invocation of a chaincode
// This chaincode will manage two accounts EMP and CPO and will transfer X units from EMP to CPO upon invoke
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Invoke called, determining function\n")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from EMP to CPO
		fmt.Printf("Function is invoke\n")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init\n")
		return t.Init(stub, function, args)
	} else if function == "delete\n" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete\n")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Run called, passing through to Invoke (same function)\n")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from EMP to CPO
		fmt.Printf("Function is invoke\n")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init\n")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete\n")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Query called, determining function\n")

	// account object as stored in blockchain
	var accountValueBytes []byte
	var err error
	var jsonResp string

	if function != "query" {
		fmt.Printf("Function is query\n")
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	//var account string // Entities
	//var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	accountKey := args[0]

	accountValueBytes, err = stub.GetState(accountKey)

	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + accountKey + "\"}"
		return nil, errors.New(jsonResp)
	}

	// return account object with values
	return accountValueBytes, nil
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s\n", err)
	}
}
