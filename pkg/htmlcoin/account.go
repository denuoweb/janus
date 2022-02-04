package htmlcoin

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

type Accounts []*btcutil.WIF

func (as Accounts) FindByHexAddress(addr string) *btcutil.WIF {
	for _, a := range as {
		acc := &Account{a}

		if addr == acc.ToHexAddress() {
			return a
		}
	}

	return nil
}

type Account struct {
	*btcutil.WIF
}

func (a *Account) ToHexAddress() string {
	// wif := (*btcutil.WIF)(a)

	keyid := btcutil.Hash160(a.SerializePubKey())
	return hex.EncodeToString(keyid)
}

var htmlcoinMainNetParams = chaincfg.MainNetParams
var htmlcoinTestNetParams = chaincfg.MainNetParams

func init() {
	htmlcoinMainNetParams.PubKeyHashAddrID = 41
	htmlcoinMainNetParams.ScriptHashAddrID = 100

	htmlcoinTestNetParams.PubKeyHashAddrID = 100
	htmlcoinTestNetParams.ScriptHashAddrID = 110
}

func (a *Account) ToBase58Address(isMain bool) (string, error) {
	params := &htmlcoinMainNetParams
	if !isMain {
		params = &htmlcoinTestNetParams
	}

	addr, err := btcutil.NewAddressPubKey(a.SerializePubKey(), params)
	if err != nil {
		return "", err
	}

	return addr.AddressPubKeyHash().String(), nil
}
