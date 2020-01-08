package restAPI

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const testNetUrl = "https://api-testnet.ergoplatform.com"
const mainNetUrl = "https://api.ergoplatform.com"

func GetCurrentHeight(testNet bool) ([]byte, error) {
	url := ""
	if testNet {
		url = testNetUrl
	} else {
		url = mainNetUrl
	}
	r, err := http.Get(url + "/blocks?limit=1")
	if err != nil {
		return nil, errors.New("can't connect to Ergo Explorer")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, nil
}

func GetBoxesFromAddress(chargeAddress string, testNet bool) ([]byte, error) {
	url := ""
	if testNet {
		url = testNetUrl
	} else {
		url = mainNetUrl
	}
	r, err := http.Get(url + "/transactions/boxes/byAddress/unspent/" + chargeAddress)
	if err != nil {
		return nil, errors.New("can't connect to Ergo Explorer")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, nil
}

func SendTx(msg []byte, testNet bool) error {
	url := ""
	if testNet {
		url = testNetUrl
	} else {
		url = mainNetUrl
	}
	url = url + "/transactions"
	r := bytes.NewReader(msg)
	resp, err := http.Post(url, "application/json", r)
	if err != nil {
		return fmt.Errorf("can't connect: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
