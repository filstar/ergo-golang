package transaction

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/jinzhu/copier"

	"github.com/zhiganov-andrew/ergo-golang/pkg/crypto"
	"github.com/zhiganov-andrew/ergo-golang/pkg/restAPI"
	"github.com/zhiganov-andrew/ergo-golang/pkg/utils"
)

const MinerErgoTree = "1005040004000e36100204a00b08cd0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798ea02d192a39a8cc7a701730073011001020402d19683030193a38cc7b2a57300000193c2b2a57301007473027303830108cdeeac93b1a57304"

type TxOutput struct {
	Address string
	Amount  int64
}

type Blocks struct {
	Items []Block `json:"items"`
	Total int64   `json:"total"`
}

type Block struct {
	Id                string `json:"id"`
	Height            int64  `json:"height"`
	TransactionsCount int64  `json:"transactionsCount"`
}

type Asset struct {
	TokenId string `json:"tokenId"`
	Amount  int64  `json:"amount"`
}

type Box struct {
	Id     string `json:"id"`
	Amount int64  `json:"value"`
	sk     string
	Assets []Asset `json:"assets"`
}

type Outputs struct {
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	Assets  []Asset `json:"assets"`
}

type Transaction struct {
	Inputs     []TransactionInput  `json:"inputs"`
	DataInputs []string            `json:"dataInputs"`
	Outputs    []TransactionOutput `json:"outputs"`
}

type TransactionOutput struct {
	ErgoTree            string         `json:"ergoTree"`
	Assets              []Asset        `json:"assets"`
	AdditionalRegisters AdditionalRegs `json:"additionalRegisters"`
	Value               int64          `json:"value"`
	CreationHeight      int64          `json:"creationHeight"`
}

type TransactionInput struct {
	BoxId         string `json:"boxId"`
	SpendingProof struct {
		ProofBytes string    `json:"proofBytes"`
		Extension  Extension `json:"extension"`
	} `json:"spendingProof"`
}

type Extension struct{}
type AdditionalRegs struct{}

func createOutputs(recipientOutputs []TxOutput, fee int64, boxesToSpend []Box, chargeAddress string) ([]Outputs, error) {
	var globalValue int64 = 0
	for _, box := range boxesToSpend {
		globalValue = globalValue + box.Amount
	}

	boxAssets := getAssetsFromBoxes(boxesToSpend)
	chargeAmount := globalValue - fee
	for _, output := range recipientOutputs {
		if output.Amount > 0 {
			chargeAmount -= output.Amount
		}
	}

	outputs := make([]Outputs, 0)
	for _, output := range recipientOutputs {
		outputs = append(outputs, Outputs{Address: output.Address, Amount: output.Amount, Assets: []Asset{}})
	}

	if chargeAmount > 0 {
		outputs = append(outputs, Outputs{Address: chargeAddress, Amount: chargeAmount, Assets: boxAssets})
	} else if chargeAmount < 0 || len(boxAssets) > 0 {
		return nil, errors.New("Not enough ERGS")
	}

	return outputs, nil
}

func getAssetsFromBoxes(boxesToSpend []Box) []Asset {
	if len(boxesToSpend) == 0 {
		return []Asset{}
	}

	allAssets := make([]Asset, 0)
	for _, box := range boxesToSpend {
		allAssets = append(allAssets, box.Assets...)
	}
	// TODO: implement
	return []Asset{}
}

func SendTransaction(recipientOutputs []TxOutput, fee int64, sk string, testNet bool) (interface{}, error) {
	chargeAddress := crypto.GetAddressFromSK(sk, testNet)
	addressBoxes, _ := restAPI.GetBoxesFromAddress(chargeAddress, testNet)
	var boxes []Box
	err := json.Unmarshal(addressBoxes, &boxes)
	if err != nil {
		return nil, errors.New("bad JSON unmarshalling")
	}

	for ind, _ := range boxes {
		boxes[ind].sk = sk
	}

	var totalAmount int64 = 0
	for _, output := range recipientOutputs {
		if output.Amount > 0 {
			totalAmount += output.Amount
		}
	}

	resolvedBoxes, err := getSolvingBoxes(boxes, totalAmount, fee)
	if err != nil {
		return nil, errors.New("Insufficient funds")
	}
	var blocks Blocks
	response, _ := restAPI.GetCurrentHeight(testNet)
	err = json.Unmarshal(response, &blocks)
	if err != nil {
		return nil, errors.New("bad JSON unmarshalling")
	}
	blockHeight := blocks.Items[0].Height

	return sendFullTransaction(recipientOutputs, fee, resolvedBoxes, chargeAddress, blockHeight, testNet)
}

func sendFullTransaction(recipientOutputs []TxOutput, fee int64, resolveBoxes []Box, chargeAddress string, blockHeight int64, testNet bool) (interface{}, error) {
	signedTransaction := formTransaction(recipientOutputs, fee, resolveBoxes, chargeAddress, blockHeight)

	msg, err := json.Marshal(&signedTransaction)
	if err != nil {
		return nil, fmt.Errorf("can't marshall tx: %w", err)
	}
	//fmt.Println("signedTransaction")
	//fmt.Println(signedTransaction)

	restAPI.SendTx(msg, testNet)

	return signedTransaction, nil
}

func formTransaction(recipientOutputs []TxOutput, fee int64, resolveBoxes []Box, chargeAddress string, blockHeight int64) Transaction {
	outputs, _ := createOutputs(recipientOutputs, fee, resolveBoxes, chargeAddress)
	signedTransaction := CreateTransaction(resolveBoxes, outputs, fee, blockHeight)

	return signedTransaction
}

func MakeErgoTree(address string) string {
	var ergoTreeBytes = []byte{0x00, 0x08, 0xcd}
	tree := append(ergoTreeBytes, crypto.GetPKFromAddress(address)...)
	return hex.EncodeToString(tree)
}

func CreateTransaction(resolveBoxes []Box, outputs []Outputs, fee int64, blockHeight int64) Transaction {
	var unsignedTransaction Transaction

	unsignedTransaction.DataInputs = make([]string, 0) // dataInputs == []

	for _, output := range outputs {
		fmt.Println(output.Address)
		treeHex := MakeErgoTree(output.Address)
		var transactionOutput TransactionOutput
		transactionOutput.AdditionalRegisters = AdditionalRegs{}
		transactionOutput.Assets = output.Assets
		transactionOutput.CreationHeight = blockHeight
		transactionOutput.ErgoTree = treeHex
		transactionOutput.Value = output.Amount
		unsignedTransaction.Outputs = append(unsignedTransaction.Outputs, transactionOutput)
	}

	if fee > 0 {
		var transactionOutput TransactionOutput
		transactionOutput.AdditionalRegisters = AdditionalRegs{}
		emptyAsset := make([]Asset, 0)
		transactionOutput.Assets = emptyAsset
		transactionOutput.CreationHeight = blockHeight
		transactionOutput.ErgoTree = MinerErgoTree
		transactionOutput.Value = fee
		unsignedTransaction.Outputs = append(unsignedTransaction.Outputs, transactionOutput)
	}

	for _, box := range resolveBoxes {
		var transactionInput TransactionInput
		transactionInput.BoxId = box.Id
		transactionInput.SpendingProof.ProofBytes = ""
		transactionInput.SpendingProof.Extension = Extension{}
		unsignedTransaction.Inputs = append(unsignedTransaction.Inputs, transactionInput)
	}

	var signedTransaction Transaction
	copier.Copy(&signedTransaction, &unsignedTransaction)

	serializeTransaction := serializeTx(unsignedTransaction)

	//fmt.Println("serializeTransaction")
	//fmt.Println("serialized:", serializeTransaction)

	//mbId := blake2b.Sum256(serializeTransaction)
	//fmt.Println(hex.EncodeToString(mbId[:]))

	for ind, _ := range signedTransaction.Inputs {
		signBytes := crypto.Sign(serializeTransaction, resolveBoxes[ind].sk)
		signedTransaction.Inputs[ind].SpendingProof.ProofBytes = hex.EncodeToString(signBytes)
	}

	return signedTransaction
}

func inputBytes(input TransactionInput) []byte {
	res, _ := hex.DecodeString(input.BoxId)
	sp := input.SpendingProof
	res = append(res, utils.IntToVlq(uint64(len(sp.ProofBytes)))...)
	proof, _ := hex.DecodeString(sp.ProofBytes)
	res = append(res, proof...)
	res = append(res, byte(0))
	return res
}

func outputBytes(output TransactionOutput) []byte {
	res := utils.IntToVlq(uint64(output.Value))
	ergoTreeBytes, _ := hex.DecodeString(output.ErgoTree)
	res = append(res, ergoTreeBytes...)
	res = append(res, utils.IntToVlq(uint64(output.CreationHeight))...)
	res = append(res, utils.IntToVlq(uint64(len(output.Assets)))...)

	res = append(res, utils.IntToVlq(uint64(0))...)

	return res
}

func serializeTx(tx Transaction) []byte {
	res := utils.IntToVlq(uint64(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		res = append(res, inputBytes(input)...)
	}

	res = append(res, utils.IntToVlq(uint64(len(tx.DataInputs)))...)

	distinctIds := make([]byte, 0)
	res = append(res, utils.IntToVlq(uint64(len(distinctIds)))...)

	res = append(res, utils.IntToVlq(uint64(len(tx.Outputs)))...)
	for _, output := range tx.Outputs {
		res = append(res, outputBytes(output)...)
	}

	return res
}

func getSolvingBoxes(boxes []Box, amount int64, fee int64) ([]Box, error) {
	var boxesCollValue int64 = 0
	var hasBoxes bool = false
	solvingBoxes := make([]Box, 0)
	sortedBoxes := sortBoxes(boxes)

	for _, box := range sortedBoxes {
		boxesCollValue = boxesCollValue + box.Amount
		solvingBoxes = append(solvingBoxes, box)

		if boxesCollValue >= amount+fee {
			hasBoxes = true
			break
		}
	}

	if !hasBoxes {
		return nil, errors.New("no solving boxes")
	}

	return solvingBoxes, nil
}

func sortBoxes(boxes []Box) []Box {
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].Amount < boxes[j].Amount
	})

	return boxes
}

type NodeBlocks struct {
	blocks       []NodeBlock
	total        uint64
	lastHeaderId string
}

type NodeBlock struct {
	Header            NodeBlockHeader       `json:"header"`
	BlockTransactions NodeBlockTransactions `json:"blockTransactions"`
	BlockExtension    NodeBlockExtension    `json:"extension"`
}

type NodeBlockExtension struct {
	HeaderId string `json:"headerId"`
}

type NodeBlockHeader struct {
	Height   uint64 `json:"height"`
	Id       string `json:"id"`
	ParentId string `json:"parentId"`
}

type NodeBlockTransactions struct {
	HeaderId     string                 `json:"headerId"`
	Transactions []NodeBlockTransaction `json:"transactions"`
	Size         uint64                 `json:"size"`
}

type NodeBlockTransaction struct {
	Id      string              `json:"id"`
	Inputs  []TransactionInput  `json:"inputs"`
	Outputs []TransactionOutput `json:"outputs"`
	Size    uint64              `json:"size"`
}
