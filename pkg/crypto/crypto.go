package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/blake2b"

	utils "ergo-golang/pkg/utils"
)

func MakeBlake2bHash(source []byte) *big.Int {
	hash := blake2b.Sum256(source)
	i := new(big.Int).SetBytes(hash[:24])
	return i
}

func GenerateRandSK() []byte {
	c := secp256k1.S256()
	randReader := rand.Reader

	params := c.Params()
	b := make([]byte, params.BitSize/8+8)
	_, err := io.ReadFull(randReader, b)
	if err != nil {
		return nil
	}

	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, new(big.Int).SetInt64(1))
	k.Mod(k, n)
	k.Add(k, new(big.Int).SetInt64(1))
	return k.Bytes()
}

func GetPKFromSK(sk string) []byte {
	curve := secp256k1.S256()
	sk_bytes, err := hex.DecodeString(sk)
	if err != nil {
		//TODO: make it adequate
		return nil
	}
	x, y := curve.ScalarBaseMult(sk_bytes)
	result := secp256k1.CompressPubkey(x, y)
	return result
}

func GetAddressFromPK(pk []byte, testNet bool) string {
	var NETWORK_TYPE byte
	const P2PK_TYPE = 1
	if testNet {
		NETWORK_TYPE = 16
	} else {
		NETWORK_TYPE = 0
	}
	prefixByte := make([]byte, 0)
	prefixByte = append(prefixByte, NETWORK_TYPE+P2PK_TYPE)
	contentBytes := pk[:]
	checksum := blake2b.Sum256(append(prefixByte, contentBytes...))
	address := append(prefixByte, contentBytes...)
	address = append(address, checksum[0:32]...)
	return base58.Encode(address[:38])
}

func GetPKFromAddress(address string) []byte {
	addressBytes := base58.Decode(address)
	return addressBytes[1:34]
}

func GetAddressFromSK(sk string, testNet bool) string {
	return GetAddressFromPK(GetPKFromSK(sk), testNet)
}

func Sign(msgbytes []byte, sk string) []byte {
	curve := secp256k1.S256()

	tryToSign := func(msgbytes []byte, sk string) []byte {
		randomToken := make([]byte, 32)
		if _, err := rand.Read(randomToken); err != nil {
			return nil
		}
		y := new(big.Int).SetBytes(randomToken[:32])

		y.Mod(y, curve.Params().N)
		x1, y1 := curve.ScalarBaseMult(y.Bytes())
		w := secp256k1.CompressPubkey(x1, y1)

		skBytes, _ := hex.DecodeString(sk)
		x2, y2 := curve.ScalarBaseMult(skBytes)
		pk := secp256k1.CompressPubkey(x2, y2)

		commitment := utils.GenCommitment(pk, w)

		hash := MakeBlake2bHash(append(commitment, msgbytes...))
		//if (c.isZero()) {
		//	return null;
		//} //TODO: make zero check on hash

		skBigint := new(big.Int).SetBytes(skBytes)

		z := new(big.Int)
		z.Mul(skBigint, hash)
		z.Add(z, y)
		z.Mod(z, curve.Params().N)

		cb := hash.Bytes()
		zb := z.Bytes()

		return append(cb, zb...)
	}

	return tryToSign(msgbytes, sk)
}

func Verify(msgBytes []byte, sigBytes []byte, pkBytes []byte) bool {
	if len(sigBytes) != 56 {
		return false
	}

	curve := secp256k1.S256()
	c := new(big.Int).SetBytes(sigBytes[:24])
	z := new(big.Int).SetBytes(sigBytes[24:56])
	pkX, pkY := secp256k1.DecompressPubkey(pkBytes)
	t := new(big.Int)
	t.Sub(curve.Params().N, c)
	tX, tY := curve.ScalarMult(pkX, pkY, t.Bytes())
	wX, wY := curve.ScalarBaseMult(z.Bytes())
	wX, wY = curve.Add(wX, wY, tX, tY)
	wb := secp256k1.CompressPubkey(wX, wY)
	commitment := utils.GenCommitment(pkBytes, wb)
	s := append(commitment, msgBytes...)
	c2 := MakeBlake2bHash(s)

	return c2.Cmp(c) == 0
}
