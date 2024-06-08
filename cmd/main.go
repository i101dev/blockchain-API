package main

import (
	"fmt"

	"github.com/i101dev/blockchain-api/utils"
)

func main() {
	fmt.Println(utils.IsFoundHost("127.0.0.1", 5000))
}
