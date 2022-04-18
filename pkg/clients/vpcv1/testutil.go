/*
Copyright 2021 The Crossplane Authors.

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

package vpcv1

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

// TypeVal converts an interface to a type (for string, *string, bool, *bool, map[string]string, *map[string]string)
//
// Params
//    value - a value. Cannot be nil
//
// Returns
//    the value of the parameter (of the appopriate type, dereferenced if a pointer), or nil
func TypeVal(value interface{}) interface{} {
	var result interface{}

	switch typed := value.(type) {
	case string:
		result = typed
	case *string:
		if typed != nil {
			result = *typed
		}
	case bool:
		result = typed
	case *bool:
		if typed != nil {
			result = *typed
		}
	case *map[string]string:
		if typed != nil {
			result = *typed
		}
	case map[string]string:
		result = typed
	}

	return result
}

// GenerateSomePermutations returns "some"  permutations (of booleans) for a given number of elements. Eg if numElems == 3,
// 9 (ie 3^2) permutations will be returned...
//
// Params
// 	  numElems - the number of elements of each array.. in the return array
//    returnSize - size of each return array (may require padding. Which can be with random vars or not...)
//    randAll - randomize the elements that "fill" the random array (we pad at the beginning). false = no randomization => deterministic values
//
// Returns
//    an array of boolean arrays (each of the given size), containing the combinations
func GenerateSomePermutations(numElems int, returnSize int, randAll bool) [][]bool {
	result := make([][]bool, 0)

	if !randAll {
		rand.Seed(int64(numElems))
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	for _, booleanCombSubset := range GeneratePermutations(numElems) {
		// Add as many variables as there are parameters
		booleanComb := make([]bool, returnSize)
		copy(booleanComb[returnSize-len(booleanCombSubset):], booleanCombSubset)

		for j := 0; j < returnSize-len(booleanCombSubset); j++ {
			booleanComb[j] = rand.Intn(1) == 1 // nolint - this is ok as we are not doing critical stuff here...)
		}

		result = append(result, booleanComb)
	}

	return result
}

// GetBinaryRep generates a binary representation in string format
//
// Params
//      i  - an integer >= 0
//      size  >= 2^i
//
// Returns
//      a string with binary representation of the integer, of length == size
func GetBinaryRep(i int, size int) string {
	result := strconv.FormatInt(int64(i), 2)

	for len(result) < size {
		result = "0" + result
	}

	return result
}

// GeneratePermutations returns all the orderings (of booleans) for a given number of elements
//
// Params
// 	  numElems - the number of elements
//
// Returns
//    an array of boolean arrays
func GeneratePermutations(numElems int) [][]bool {
	result := make([][]bool, 0)

	for i := 0; i < int(math.Pow(2, float64(numElems))); i++ {
		str := GetBinaryRep(i, numElems)

		boolArray := make([]bool, numElems)
		boolArrayIdx := len(boolArray) - 1
		for j := len(str) - 1; j >= 0; j-- {
			boolArray[boolArrayIdx] = (str[j] == '1')

			boolArrayIdx--
		}

		result = append(result, boolArray)
	}

	return result
}
