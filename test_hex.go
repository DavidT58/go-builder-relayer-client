package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	data := []byte{0xa9, 0x05, 0x9c, 0xbb}
	
	// Test Bytes2Hex
	hex1 := common.Bytes2Hex(data)
	fmt.Printf("Bytes2Hex: %s\n", hex1)
	
	// Test hexutil.Encode
	hex2 := common.Bytes2Hex(data)
	fmt.Printf("Should be: 0x%s\n", hex2)
}
