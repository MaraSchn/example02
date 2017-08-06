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
	"errors"
	"fmt"
	"strconv"
	"encoding/json"


	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// transactions with metadata, always stored in accounts
type Transaction struct{
	session_id string `json:"session_id"`
	cpo string `json:"cpo"`
	emp string `json:"emp"`
	product string `json:"product"`
	evse_id string `json:"evse_id"`
	user_id string `json:"user_id"`
	timestamp string `json:"timestamp"`
	charging_duration float32 `json:"charging_duration"`
	charged_energy float32 `json:"charged_energy"`
	price_per_unit float32 `json:"price_per_unit"`
	value_brutto int `json:"value_brutto"`
}

// every transaction belongs to the receiving and the transmitting account
// an acount contains the total account balance and the history of transactions affecting the account
type Account struct{
	// balance of the account including taxation etc.
	balance_brutto int `json:"balance_brutto"`
	//array of transactions an account had
	//transactions []Transaction `json:"transactions"`
}




func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Maraike said: Init called, initializing chaincode")

	return nil, nil
}


// ============================================================================================================================
// Transaction: payment of X euro cents from EMP to CPO
// ============================================================================================================================
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running invoke")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting, in this order: EMP, CPO, transaction value")
	}

	// entity keys / identifiers
	var emp_key, cpo_key string
	// entities
	var emp_account, cpo_account Account
	// transaction value
	var tranaction_value int
	// updated entities, to be written to blockchain
	var emp_account_bytes, cpo_account_bytes []byte

	var err error

	// set involved EMP and CPO from function call
	emp_key = args[0]
	cpo_key = args[1]

	// TODO: REMOVE OVERWRITING ACCOUNTS WHEN DEPLOYING THIS BRANCH FOR THE SECOND TIME

	/*t.delete(stub, emp_key)
	t.delete(stub, cpo_key)
	*/

	new_account := Account{}
	new_account.balance_brutto = 0
	jsonAsBytes, _ := json.Marshal(new_account)
	err = stub.PutState(emp_key, jsonAsBytes)
	err = stub.PutState(cpo_key, jsonAsBytes)


	// load EMPs account
	emp_account, err = t.getOrCreateNewAccount(stub, emp_key)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Account balance of %s prior transaction is %d €. ", emp_key, emp_account.balance_brutto/100)

	// load CPOs account
	cpo_account, err = t.getOrCreateNewAccount(stub, cpo_key)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Account balance of %s prior transaction is %d €. ", cpo_key, cpo_account.balance_brutto/100)

	// Calculate the new total account balances for EMP and CPO
	tranaction_value, err = strconv.Atoi(args[2])
	if err != nil {
		return nil, err
	}
	emp_account.balance_brutto = emp_account.balance_brutto - tranaction_value
	cpo_account.balance_brutto = cpo_account.balance_brutto + tranaction_value

	//fmt.Printf("EMP_balance = %d, CPO_balance = %d\n", EMP_balance, CPO_balance)
	fmt.Printf("Account balance of %s after transaction is %d €. ", emp_key, emp_account.balance_brutto/100)
	fmt.Printf("Account balance of %s after transaction is %d €. ", cpo_key, cpo_account.balance_brutto/100)

	// write the updated EMP account back to the ledger
	emp_account_bytes, err = json.Marshal(emp_account)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(emp_key, emp_account_bytes)
	if err != nil {
		return nil, err
	}

	// write the updated CPO acccount back to the ledger
	cpo_account_bytes, err = json.Marshal(cpo_account)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(cpo_key, cpo_account_bytes)
	if err != nil {
		return nil, err
	}

	// function ran through without errors
	return nil, nil
}


// ============================================================================================================================
// read an account from the blockchain; if it doesn't exist yet, create a new one beforehand; return the account
// ============================================================================================================================
func (t *SimpleChaincode) getOrCreateNewAccount(stub shim.ChaincodeStubInterface, args []string) (Account, error) {
	var account_key, jsonResp string
	// account object as stored in blockchain
	var account_value_bytes []byte
	var err error
	// empty account template
	var account Account
	account = Account{}

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	account_key = args[0]
	account_value_bytes, err = stub.GetState(account_key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + account_key + "\"}"
		return nil, errors.New(jsonResp)
	}

	// if loading account returned no bytes, no account exists
	// --> create an account and load new accounts object bytes
	if account_value_bytes == nil {
		fmt.Printf("%s has no account in the ledger yet.", account_key)
		account.balance_brutto = 0
		account_value_bytes, err = json.Marshal(account)
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to marshal json for new account " + account_key + "\"}"
			return nil, errors.New(jsonResp)
		}

		err = stub.PutState(account_key, account_value_bytes)
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to put state for new account " + account_key + "\"}"
			return nil, errors.New(jsonResp)
		}
	}

	// fill account template with values read from blockchain
	json.Unmarshal(account_value_bytes, &account)

	// return account object with values
	return account, nil
}


// ============================================================================================================================
// Deletes an entity from state
// ============================================================================================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running delete")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	account_key := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(account_key)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}

// Invoke callback representing the invocation of a chaincode
// This chaincode will manage two accounts EMP and CPO and will transfer X units from EMP to CPO upon invoke
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Invoke called, determining function")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from EMP to CPO
		fmt.Printf("Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t* SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Run called, passing through to Invoke (same function)")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from EMP to CPO
		fmt.Printf("Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Query called, determining function")

	if function != "query" {
		fmt.Printf("Function is query")
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var account string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	account = args[0]

	// Get the state from the ledger
	account_balance_bytes, err := stub.GetState(account)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + account + "\"}"
		return nil, errors.New(jsonResp)
	}

	if account_balance_bytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + account + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + account + "\",\"Amount\":\"" + string(account_balance_bytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return account_balance_bytes, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
