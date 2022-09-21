package main

import (
	"fmt"
	"github.com/zhiganov-andrew/ergo-golang/pkg/crypto"
	"github.com/zhiganov-andrew/ergo-golang/pkg/transaction"
)

func main() {
	//crypto.GetSKWithMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", "TREZOR")
	bsk, sk := crypto.GetSKWithMnemonic("van impose hub swift ladder example celery two omit soft mobile very again pause satoshi", "")
	fmt.Println(crypto.GetAddressFromSK(bsk, false))
	fmt.Println(crypto.GetAddressFromSK(sk, false))
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
