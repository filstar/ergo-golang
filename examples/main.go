package main

import (
	"github.com/zhiganov-andrew/ergo-golang/pkg/crypto"
	"github.com/zhiganov-andrew/ergo-golang/pkg/transaction"
)

func main() {
	crypto.GetSKWithMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", "TREZOR")
	return
	outputs := []transaction.TxOutput{
		{
			Address: "3WywqB19PtvCTogmGYRX3eKad2iiCjNJkeYGFiSjVEGRbFUJ4dAA",
			Amount:  5000000000,
		},
		{
			Address: "3WywqB19PtvCTogmGYRX3eKad2iiCjNJkeYGFiSjVEGRbFUJ4dAA",
			Amount:  1000000000,
		},
	}

	transaction.SendTransaction(outputs, 1000000, "3WywqB19PtvCTogmGYRX3eKad2iiCjNJkeYGFiSjVEGRbFUJ4dAA", true)
}
