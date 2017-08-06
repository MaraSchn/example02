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

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}
//TODO: keine Argumente in der Init-Funktion übergeben, alles einfach ausblenden???
//TODO: Prüfung Umbenennung A = EMP, B = CPO, Aval = EMP_balance, Bval = CPO_balance, X= transaction_value
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Maraike said: Init called, initializing chaincode")

	/*var EMP, CPO string    // Entities
	var EMP_balance, CPO_balance int // Asset holdings
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	EMP = args[0]
	EMP_balance, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	CPO = args[2]
	CPO_balance, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	fmt.Printf("EMP_balance = %d, CPO_balance = %d\n", EMP_balance, CPO_balance)

	// Write the state to the ledger
	err = stub.PutState(EMP, []byte(strconv.Itoa(EMP_balance)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(CPO, []byte(strconv.Itoa(CPO_balance)))
	if err != nil {
		return nil, err
	}

	return nil, nil */
}

// Transaction makes payment of X units from EMP to CPO
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running invoke")

	var EMP, CPO string    // Entities
	var EMP_balance, CPO_balance int // Account balance
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	EMP = args[0]
	CPO = args[1]

	// var EMP_balance_bytes

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	EMP_balance_bytes, err := stub.GetState(EMP)
	if err != nil {

		return nil, errors.New("Failed to get state")
	}
	if EMP_balance_bytes == nil {
		fmt.Printf("%s has no account in the ledger yet.", EMP)
		err = stub.PutState(EMP, []byte(strconv.Itoa(0)))

		EMP_balance_bytes, err = stub.GetState(EMP)
		if err != nil {fmt.Printf("Failed to load new account.")}

		// return nil, errors.New("Entity not found")
	}


	EMP_balance, _ = strconv.Atoi(string(EMP_balance_bytes))
	fmt.Printf("Account balance of %s prior transaction is %d €. ", EMP, EMP_balance/100 )






	CPO_balance_bytes, err := stub.GetState(CPO)
	if err != nil {

		return nil, errors.New("Failed to get state")
	}
	if CPO_balance_bytes == nil {

		fmt.Printf("%s has no account in the ledger yet.", CPO)
		err = stub.PutState(CPO, []byte(strconv.Itoa(0)))

		CPO_balance_bytes, err = stub.GetState(CPO)
		if err != nil {fmt.Printf("Failed to load new account.")}

		// return nil, errors.New("Entity not found")
	}



	CPO_balance, _ = strconv.Atoi(string(CPO_balance_bytes))

	fmt.Printf("Account balance of %s prior transaction is %d €. ", CPO, CPO_balance/100 )







	// Perform the execution
	X, err = strconv.Atoi(args[2])
	EMP_balance = EMP_balance - X
	CPO_balance = CPO_balance + X
	//fmt.Printf("EMP_balance = %d, CPO_balance = %d\n", EMP_balance, CPO_balance)
	fmt.Printf("Account balance of %s after transaction is %d €. ", EMP, EMP_balance/100 )
	fmt.Printf("Account balance of %s after transaction is %d €. ", CPO, CPO_balance/100 )

	// Write the state back to the ledger
	err = stub.PutState(EMP, []byte(strconv.Itoa(EMP_balance)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(CPO, []byte(strconv.Itoa(CPO_balance)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running delete")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	EMP := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(EMP)
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
