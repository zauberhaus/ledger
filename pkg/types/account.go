package types

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math/big"
	"strconv"
	"strings"
)

const (
	AllAccounts Account = ""
)

type AccountInfo struct {
	Account  Account
	Customer string
	Asset    Asset
}

type Account string

func NewAccount(customer string, asset Asset) (Account, error) {
	hash := crc32.New(crc32.IEEETable)
	hash.Write([]byte(customer))
	account := hex.EncodeToString(hash.Sum([]byte(asset)))

	chk := Account(account + "00").Checksum()

	return Account(fmt.Sprintf("%v%02d", account, chk)), nil

}

func (a Account) String() string {
	return string(a)
}

func (a Account) Empty() bool {
	return len(a) == 0
}

func (a Account) Check() bool {
	if len(a) < 3 {
		return false
	}

	account := string(a)

	expected, err := strconv.ParseUint(account[len(account)-2:], 10, 8)
	if err != nil {
		return false
	}

	got := a.Checksum()

	return got == uint8(expected)
}

func (a Account) Checksum() uint8 {
	if len(a) < 3 {
		return 0
	}

	account := string(a[:len(a)-2])
	account = strings.ToUpper(account)

	account = strings.ReplaceAll(account, "A", "1")
	account = strings.ReplaceAll(account, "B", "2")
	account = strings.ReplaceAll(account, "C", "3")
	account = strings.ReplaceAll(account, "D", "4")
	account = strings.ReplaceAll(account, "E", "5")
	account = strings.ReplaceAll(account, "F", "6")
	account = strings.ReplaceAll(account, "G", "7")
	account = strings.ReplaceAll(account, "H", "8")
	account = strings.ReplaceAll(account, "I", "9")
	account = strings.ReplaceAll(account, "J", "1")
	account = strings.ReplaceAll(account, "K", "2")
	account = strings.ReplaceAll(account, "L", "3")
	account = strings.ReplaceAll(account, "M", "4")
	account = strings.ReplaceAll(account, "N", "5")
	account = strings.ReplaceAll(account, "O", "6")
	account = strings.ReplaceAll(account, "P", "7")
	account = strings.ReplaceAll(account, "Q", "8")
	account = strings.ReplaceAll(account, "R", "9")
	account = strings.ReplaceAll(account, "S", "2")
	account = strings.ReplaceAll(account, "T", "3")
	account = strings.ReplaceAll(account, "U", "4")
	account = strings.ReplaceAll(account, "V", "5")
	account = strings.ReplaceAll(account, "W", "6")
	account = strings.ReplaceAll(account, "X", "7")
	account = strings.ReplaceAll(account, "Y", "8")
	account = strings.ReplaceAll(account, "Z", "9")

	val := new(big.Int)
	val.SetString(account, 10)

	mod := new(big.Int)
	mod.SetInt64(97)

	chkSum := new(big.Int)
	chkSum.Mod(val, mod)

	return uint8(97 - chkSum.Int64())
}
