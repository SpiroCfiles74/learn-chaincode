/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	err = stub.CreateTable("C", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Asset", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Owner", Type: shim.ColumnDefinition_BYTES, Key: false},
	})
	if err != nil {
		return nil, err
	}

	err = stub.CreateTable("M", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Name", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "date", Type: shim.ColumnDefinition_BYTES, Key: false},
	})
	if err != nil {
		return nil, err
	}

	err = stub.CreateTable("A", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Member", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "family", Type: shim.ColumnDefinition_BYTES, Key: false},
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "delete" {
		return t.Delete(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	//test ops
	/*ID, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
	}
	fmt.Println(ID)
	msg := stub.GetStringArgs()
	fmt.Println(msg)*/
	key := args[0]

	keysIter, err := stub.RangeQueryState(args[0], args[1])
	if err != nil {
		return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
	}
	defer keysIter.Close()

	var keys []string
	for keysIter.HasNext() {
		key, _, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		keys = append(keys, key)
	}

	/*resultindex := "namenum"
	resultindexkey, err := stub.createcompositkey(resultindex, []string{args[0],args[1]})
	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	stub.PutState(resultindexkey, value)*/

	table, err := GetTable("C")
	if err != nil {
		return nil, errors.New(err.Error())
	}

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: key}}
	columns = append(columns, col1)
	Row, err := stub.GetRow("C", columns)
	if err != nil {
		return nil, err
	}

	err = stub.DeleteTable("C")
	if err != nil {
		return nil, err
	}

	table2, err := GetTable("C")
	if err != nil {
		return nil, errors.New(err.Error())
	}
	fmt.Printf(table2)

	state, err := InsertRow("M", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: Name}},
			&shim.Column{Value: &shim.Column_String_{String_: date}},
			&shim.Column{Value: &shim.Column_String_{String_: family}},
			&shim.Column{Value: &shim.Column_String_{String_: Member}},
		}})
	if !state && err != nil {
		return nil, err
	}
	tablem, err := GetTable("M")

	if err != nil {
		return nil, errors.New(err.Error())
	}
	fmt.Printf(tablem)

	/*var ca stub.certificate
	var sign stub.signature
	var msg []byte

	ops, err = stub.VerifySignature(ca, sign, msg)
	if err != nil {
		return nil, err
	}*/

	/*err = stub.SetEvent(label, data)
	if err != nil {
		return nil, err
	}*/

	return nil, errors.New("Received unknown function invocation")
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

// write - invoke function to write key/value pair
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Delete - remove a key/value pair from state
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	record := args[0]
	err := stub.DelState(record) //remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}
