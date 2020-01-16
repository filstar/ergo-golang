package utils

import (
	"encoding/hex"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/blake2b"
)

func CheckAddressValidity(address string) bool {
	addressBytes := base58.Decode(address)
	size := len(addressBytes)
	if size < 8 {
		return false
	}
	script := addressBytes[:(size - 4)]
	checksum := addressBytes[(size - 4):size]
	calculatedChecksum := blake2b.Sum256(script)
	return hex.EncodeToString(calculatedChecksum[0:4]) == hex.EncodeToString(checksum)
}

func GenCommitment(pk []byte, w []byte) []byte {
	prefix, _ := hex.DecodeString("010027100108cd")
	postfix, _ := hex.DecodeString("73000021")
	return append(append(append(prefix, pk...), postfix...), w...)
}

func IntToVlq(num uint64) []byte {
	res := make([]byte, 0)
	var x uint64 = num
	var r byte
	for (x / 128) > 0 {
		r = byte(x & 0x7F)
		x = x / 128
		res = append(res, (r | 0x80))
	}
	r = byte(x & 0x7F)
	res = append(res, r)
	return res
}

func LongToVlq(long *big.Int) []byte {
	res := make([]byte, 0)
	var x = long.Set(long)
	var temp = new(big.Int)
	var temp2 = new(big.Int)
	var temp3 = new(big.Int)
	var r byte
	var rr []byte
	var xCompare = new(big.Int)
	for (xCompare.Div(x, temp.SetUint64(128))).Cmp(temp2.SetUint64(0)) > 0 {
		rr = xCompare.And(x, temp3.SetUint64(0x7F)).Bytes()
		if len(rr) > 0 {
			r = rr[0]
		} else {
			r = 0
		}

		x = x.Div(x, temp.SetUint64(128))
		res = append(res, (r | 0x80))
	}
	rr = xCompare.And(x, temp3.SetUint64(0x7F)).Bytes()
	if len(rr) > 0 {
		r = rr[0]
	} else {
		r = 0
	}
	res = append(res, r)
	return res
}

//func OutputBytes(out) {
//	res := IntToVlq(out.value)
//	ergotree, _ := hex.DecodeString(out.ergoTree)
//	res = append(res, ergotree)
//}

//export const outputBytes = (out) => {
//let res = intToVlq(out.value);
//res = Buffer.concat([res, Buffer.from(out.ergoTree, 'hex')]);
//res = Buffer.concat([res, intToVlq(out.creationHeight)]);
//
//res = Buffer.concat([res, intToVlq(out.assets.length)]);
//const k = out.additionalRegisters.length;
//res = Buffer.concat([res, intToVlq(k)]);
//return res;
//};
