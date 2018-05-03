// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merkle

import (
	"fmt"
)

// Prints the in-memory children recursively.
func PrintIAVLNode(node *IAVLNode) {
	fmt.Println("==== NODE")
	if node != nil {
		printIAVLNode(node, 0)
	}
	fmt.Println("==== END")
}

func printIAVLNode(node *IAVLNode, indent int) {
	indentPrefix := ""
	for i := 0; i < indent; i++ {
		indentPrefix += "    "
	}

	if node.rightNode != nil {
		printIAVLNode(node.rightNode, indent+1)
	} else if node.rightHash != nil {
		fmt.Printf("%s    %X\n", indentPrefix, node.rightHash)
	}

	fmt.Printf("%s%v:%v\n", indentPrefix, node.key, node.height)

	if node.leftNode != nil {
		printIAVLNode(node.leftNode, indent+1)
	} else if node.leftHash != nil {
		fmt.Printf("%s    %X\n", indentPrefix, node.leftHash)
	}

}

func maxInt8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
