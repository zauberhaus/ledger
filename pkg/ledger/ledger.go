package ledger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/crc64"
	"net/http"
	"strings"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/shopspring/decimal"

	"github.com/ec-systems/core.ledger.server/pkg/client"
	"github.com/ec-systems/core.ledger.server/pkg/ledger/index"
	"github.com/ec-systems/core.ledger.server/pkg/logger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
)

const (
	IDLength = 6
	Version  = uint16(1)

	AccountNotFoundError = 1
	TooManyAccountsError = 2
	NotEnoughAssetsError = 3
	BadRequestError      = http.StatusBadRequest
	NotFoundError        = http.StatusNotFound
	NotAcceptable        = http.StatusNotAcceptable
	InternalError        = http.StatusInternalServerError
)

type Ledger struct {
	client   *client.Client
	readOnly bool
	overdraw bool
	multi    bool

	assets   types.Assets
	statuses types.Statuses

	format types.Format

	collectors []types.MetricsCollector
}

func New(client *client.Client, options ...LedgerOption) *Ledger {
	ledger := &Ledger{
		client: client,
		format: types.JSON,
	}

	for _, option := range options {
		if option != nil {
			option.Set(ledger)
		}
	}

	if ledger.readOnly {
		logger.Info("Run ledger in read-only mode")
	}

	return ledger
}

func (l *Ledger) SupportedAssets() types.Assets {
	return l.assets
}

func (l *Ledger) SupportedStatus() types.Statuses {
	return l.statuses
}

func (l *Ledger) Assets(ctx context.Context) ([]types.Asset, error) {
	assets := []types.Asset{}

	err := l.ForEach(ctx, index.Asset.Assets(), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		assets = append(assets, tx.Asset)
		return true, nil
	})

	if err != nil {
		return nil, NewError(InternalError, "failed to load list of assets: %v", err)
	}

	return assets, nil
}

func (l *Ledger) AccountInfo(ctx context.Context, account types.Account) (*types.AccountInfo, error) {
	if account == "" {
		return nil, NewError(BadRequestError, "account is mandatory")
	}

	entries, err := l.client.Scan(ctx, string(index.Account.Key(account)), 1, false)
	if err != nil && strings.HasPrefix(err.Error(), "cant") {
		return nil, NewError(InternalError, "get account info failed: %v", err)
	}

	if len(entries) > 0 {
		tx := &Transaction{}
		err = tx.Parse(entries[0])
		if err != nil {
			return nil, err
		}

		return &types.AccountInfo{
			Account: tx.Account,
			Holder:  tx.Holder,
			Asset:   tx.Asset,
		}, nil
	}

	return nil, nil
}

func (l *Ledger) Accounts(ctx context.Context, holder string, asset types.Asset) ([]types.Account, error) {
	accounts := []types.Account{}

	if holder == "" {
		return nil, NewError(BadRequestError, "accounts: holder is mandatory")
	}

	err := l.ForEach(ctx, index.Holder.Accounts(holder, asset), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		if holder == tx.Holder {
			accounts = append(accounts, tx.Account)
		} else {
			logger.Errorf("accounts: unexpected holder %v!=%v in %v", holder, tx.Holder, tx.tx)
		}
		return true, nil
	})

	if err != nil {
		return nil, NewError(InternalError, "list accounts failed: %v", err)
	}

	return accounts, nil
}

func (l *Ledger) Holders(ctx context.Context, f func(holder string, account types.Account, asset types.Asset) (bool, error)) error {

	err := l.ForEach(ctx, index.Holder.All(), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		return f(tx.Holder, tx.Account, tx.Asset)
	})

	if err != nil {
		return NewError(InternalError, "list accounts failed: %v", err)
	}

	return nil
}

func (l *Ledger) Balance(ctx context.Context, holder string, asset types.Asset, account types.Account, status types.Status) (map[types.Asset]*types.Balance, error) {
	assets := map[types.Asset]*types.Balance{}
	var accounts []types.Account

	if holder == "" {
		return nil, NewError(BadRequestError, "balance: holder is mandatory")
	}

	if account.Empty() {
		acc, err := l.Accounts(ctx, holder, asset)
		if err != nil {
			return nil, err
		}

		accounts = acc
	} else {
		accounts = []types.Account{account}
	}

	for _, a := range accounts {
		err := l.ForEachInSet(ctx, index.Transaction.Scan(a), false, func(ctx context.Context, tx *Transaction) (bool, error) {

			if holder == tx.Holder {
				balance, ok := assets[tx.Asset]
				if !ok {
					balance = types.NewBalance(status)
					assets[tx.Asset] = balance
				}

				balance.Add(tx.Account, tx.Amount, tx.Status)
			} else {
				return false, fmt.Errorf("balance: unexpected holder %v!=%v in %v", holder, tx.Holder, tx.tx)
			}

			return true, nil
		})

		if err != nil {
			return nil, err
		}
	}

	return assets, nil
}

func (l *Ledger) AssetBalance(ctx context.Context, asset types.Asset) (map[types.Asset]decimal.Decimal, error) {
	if !l.readOnly {
		return nil, NewError(NotAcceptable, "not a read-only instance")
	}

	var assets []types.Asset

	if asset == types.AllAssets {
		tmp, err := l.Assets(ctx)
		if err != nil {
			return nil, err
		}
		assets = tmp
	} else {
		assets = []types.Asset{asset}
	}

	balances := map[types.Asset]decimal.Decimal{}

	for _, a := range assets {
		err := l.ForEachInSet(ctx, string(index.AssetTx.Key(a)), false, func(ctx context.Context, tx *Transaction) (bool, error) {
			balance, ok := balances[tx.Asset]
			if !ok {
				balance = decimal.Zero
			}

			balances[tx.Asset] = balance.Add(tx.Amount)
			return true, nil
		})

		if err != nil {
			return nil, NewError(InternalError, "failed to load list of assets: %v", err)
		}
	}

	return balances, nil
}

func (l *Ledger) Transactions(ctx context.Context, holder string, asset types.Asset, account types.Account, f func(context.Context, *Transaction) (bool, error)) error {
	if holder == "" {
		return NewError(BadRequestError, "holder is mandatory")
	}

	var accounts []types.Account
	if account.Empty() {
		acc, err := l.Accounts(ctx, holder, asset)
		if err != nil {
			return err
		}

		accounts = acc
	} else {
		accounts = []types.Account{account}
	}

	for _, a := range accounts {
		err := l.ForEachInSet(ctx, index.Transaction.Scan(a), false, func(ctx context.Context, tx *Transaction) (bool, error) {
			if holder != "" && holder != tx.Holder {
				return false, NewError(BadRequestError, "invalid holder %v in tx %v (%v)", tx.Holder, tx.ID, holder)
			}

			if asset != types.AllAssets && asset != tx.Asset {
				return false, NewError(BadRequestError, "invalid asset %v in tx %v (%v)", tx.Asset, tx.ID, asset)
			}

			if account != types.AllAccounts && account != tx.Account {
				return false, NewError(BadRequestError, "invalid account %v in tx %v (%v)", tx.Account, tx.ID, account)
			}

			return f(ctx, tx)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Ledger) Orders(ctx context.Context, holder string, f func(context.Context, *Transaction) (bool, error)) error {
	if holder == "" {
		return NewError(BadRequestError, "holder is mandatory")
	}

	return l.ForEach(ctx, index.Order.Orders(holder), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		if holder != "" && holder != tx.Holder {
			return false, NewError(BadRequestError, "invalid holder %v in tx %v (%v)", tx.Holder, tx.ID, holder)
		}

		return f(ctx, tx)
	})
}

func (l *Ledger) OrderItems(ctx context.Context, holder string, order string, item string, f func(context.Context, *Transaction) (bool, error)) error {
	if order == "" {
		return NewError(BadRequestError, "holder is mandatory")
	}

	return l.ForEachInSet(ctx, index.OrderItem.Scan(order), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		if holder != "" && holder != tx.Holder {
			return false, NewError(BadRequestError, "invalid holder %v in tx %v (%v)", tx.Holder, tx.ID, holder)
		}

		if order != "" && order != tx.Order {
			return false, NewError(BadRequestError, "invalid order %v in tx %v (%v)", tx.Order, tx.ID, order)
		}

		if item != "" && item != tx.Item {
			return false, NewError(BadRequestError, "invalid item %v in tx %v (%v)", tx.Item, tx.ID, item)
		}

		return f(ctx, tx)
	})
}

func (l *Ledger) Get(ctx context.Context, transaction types.ID) (*Transaction, error) {
	entry, err := l.client.Get(ctx, string(index.Key.Key(transaction)))
	if err != nil {
		return nil, err
	}

	tx := &Transaction{}
	err = tx.Parse(entry)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (l *Ledger) Status(ctx context.Context, in *Transaction, status types.Status) (*Transaction, error) {
	if l.readOnly {
		return nil, NewError(NotFoundError, "read-only instance")
	}

	tx, err := l.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if tx.Holder != in.Holder || tx.Asset != in.Asset || tx.Account != in.Account || tx.ID != in.ID {
		return nil, fmt.Errorf("invalid holder/asset/account/id combination")
	}

	if tx.Status != status {
		now := time.Now()
		tx.Modified = &now
		tx.Status = status

		data, err := tx.Bytes(l.format)
		if err != nil {
			return nil, NewError(InternalError, "marshal transaction failed: %v", err)
		}

		txID, err := l.client.Set(ctx, index.Key.Key(tx.ID), data)
		if err != nil {
			return nil, err
		}

		tx.tx = txID
	}

	return tx, nil
}

func (l *Ledger) CreateTx(ctx context.Context, holder string, asset types.Asset, amount decimal.Decimal, options ...TransactionOption) (*Transaction, error) {
	if l.readOnly {
		return nil, NewError(NotFoundError, "read-only instance")
	}

	if amount.IsZero() {
		return nil, NewError(BadRequestError, "transaction for holder %v with 0 %v", holder, asset)
	}

	id, err := l.NewID()
	if err != nil {
		return nil, err
	}

	tx := &Transaction{
		ID:     id,
		Holder: holder,

		Status: types.Created,
		Asset:  asset,
		Amount: amount,
	}

	for _, option := range options {
		if option != nil {
			option.Set(tx)
		}
	}

	if tx.Account == "" {
		accounts, err := l.Accounts(ctx, holder, asset)
		if err != nil {
			return nil, err
		}

		if len(accounts) == 0 {
			account, err := l.NewAccount(ctx, holder, asset)
			if err != nil {
				return nil, err
			}

			tx.Account = account
		} else if len(accounts) == 1 {
			tx.Account = accounts[0]
		} else if tx.Amount.IsZero() || tx.Amount.IsPositive() {
			return nil, NewError(TooManyAccountsError, "more than one account found for holder %v", holder)
		}
	} else {
		info, err := l.AccountInfo(ctx, tx.Account)
		if err != nil {
			return nil, NewError(InternalError, "failed to read account %v info", tx.Account)
		}

		if !l.multi && info == nil {
			return nil, NewError(AccountNotFoundError, "account %v not found", tx.Account)

		}

		if info != nil {
			if info.Holder != tx.Holder {
				return nil, NewError(BadRequestError, "invalid holder %v for account %v (%v)", tx.Holder, tx.Account, info.Holder)
			}

			if info.Asset != tx.Asset {
				return nil, NewError(BadRequestError, "invalid asset %v for account %v (%v)", tx.Asset, tx.Account, info.Asset)
			}
		}
	}

	if tx.Amount.IsNegative() && !l.overdraw {
		negAmount := amount.Neg()

		balances, err := l.Balance(ctx, holder, asset, tx.Account, types.Created)
		if err != nil {
			return nil, NewError(InternalError, "failed to get holder %v balance: %v", holder, err)
		}

		balance, ok := balances[asset]
		if !ok {
			return nil, NewError(NotFoundError, "no %v account found for holder %v", asset, holder)
		}

		if balance.Sum.LessThan(negAmount) {
			return nil, NewError(NotEnoughAssetsError, "balance too low to remove %v %v for holder %v", asset, negAmount, holder)
		}

		for k, v := range balance.Accounts {
			if !v.Sum.LessThan(negAmount) {
				tx.Account = k
				break
			}
		}

		if tx.Account == "" {
			return nil, NewError(NotFoundError, "no account found with enough balance to remove %v %v for holder %v", asset, negAmount, holder)
		}
	}

	if ok := tx.Account.Check(); !ok {
		return nil, NewError(BadRequestError, "invalid checksum for account %v", tx.Account)
	}

	ops, key, err := l.CreateOperations(tx)
	if err != nil {
		return nil, err
	}

	txID, err := l.client.Exec(ctx, ops...)

	tx.tx = txID
	tx.key = key

	for _, c := range l.collectors {
		c.Add(tx.Asset, tx.Amount)
	}

	return tx, err
}

func (l *Ledger) Add(ctx context.Context, holder string, asset types.Asset, amount decimal.Decimal, options ...TransactionOption) (*Transaction, error) {
	if amount.IsZero() || amount.IsNegative() {
		return nil, NewError(BadRequestError, "can't add %v %v", asset, amount)
	}

	return l.CreateTx(ctx, holder, asset, amount, options...)
}

func (l *Ledger) Remove(ctx context.Context, holder string, asset types.Asset, amount decimal.Decimal, options ...TransactionOption) (*Transaction, error) {
	if amount.IsZero() || amount.IsNegative() {
		return nil, NewError(BadRequestError, "can't remove %v %v", asset, amount.Neg())
	}

	return l.CreateTx(ctx, holder, asset, amount.Neg(), options...)
}

func (l *Ledger) Cancel(ctx context.Context, holder string, asset types.Asset, account types.Account, transaction types.ID) (*Transaction, error) {
	if l.readOnly {
		return nil, NewError(NotFoundError, "read-only instance")
	}

	tx, err := l.Get(ctx, transaction)
	if err != nil {
		return nil, NewError(InternalError, "cant read transaxtion %v: %v", transaction, err)
	}

	if tx.Holder != holder && tx.Asset != asset && tx.Account != account {
		return nil, NewError(BadRequestError, "inconsistent holder/account/transaction combination (%v/%v/%v)", holder, account, transaction)
	}

	now := time.Now()
	tx.Modified = &now

	ops, cancel, err := l.CancelOperations(tx)
	if err != nil {
		return nil, err
	}

	txID, err := l.client.Exec(ctx, ops...)

	cancel.tx = txID

	for _, c := range l.collectors {
		c.Add(tx.Asset, tx.Amount)
	}

	return cancel, err
}

func (l *Ledger) History(ctx context.Context, id types.ID, f func(ctx context.Context, tx *Transaction) (bool, error)) error {
	return l.client.History(ctx, index.Key.ID(id), func(ctx context.Context, e *schema.Entry) (bool, error) {
		if e.Value[0] == 0 {
			e, err := l.client.GetAt(ctx, string(e.Key), e.Tx)
			if err != nil {
				return true, NewError(InternalError, "failed to read the transaction (%v): %v", e.Tx, string(e.Value))
			}

			tx := &Transaction{}
			err = tx.Parse(e)
			if err != nil {
				return true, NewError(InternalError, "failed to parse the transaction (%v): %v", string(e.Value), err)
			}

			return f(ctx, tx)
		} else {
			tx := &Transaction{}
			err := tx.Parse(e)
			if err != nil {
				return true, NewError(InternalError, "failed to parse the transaction (%v): %v", string(e.Value), err)
			}

			return f(ctx, tx)
		}
	})
}

func (l *Ledger) ForEach(ctx context.Context, prefix string, desc bool, f func(context.Context, *Transaction) (bool, error)) error {
	return l.client.ScanAll(ctx, prefix, false, func(ctx context.Context, i int, e *schema.Entry) (bool, error) {
		tx := &Transaction{}
		err := tx.Parse(e)
		if err != nil {
			return true, NewError(InternalError, "failed to parse the transaction (%v): %v", err, string(e.Value))
		}

		return f(ctx, tx)
	})
}

func (l *Ledger) ForEachInSet(ctx context.Context, prefix string, desc bool, f func(context.Context, *Transaction) (bool, error)) error {
	return l.client.ScanSet(ctx, prefix, false, func(ctx context.Context, e *schema.ZEntry) (bool, error) {
		tx := &Transaction{}
		err := tx.Parse(e.Entry)
		if err != nil {
			return true, NewError(InternalError, "failed to parse the transaction (%v): %v", err, string(e.Entry.Value))
		}

		return f(ctx, tx)
	})
}

func (l *Ledger) NewID() (types.ID, error) {
	token := make([]byte, 16)
	n, err := rand.Read(token)

	if err != nil {
		return types.ZeroID, err
	}

	if n != 16 {
		return types.ZeroID, fmt.Errorf("can't generate transaction id")
	}

	return types.NewID(token), nil
}

func (l *Ledger) NewAccount(ctx context.Context, holder string, asset types.Asset) (types.Account, error) {
	cnt := uint8(0)

	for {
		hash := crc64.New(crc64.MakeTable(crc64.ECMA))
		hash.Write([]byte(holder))
		hash.Write([]byte(asset))
		crc := hash.Sum([]byte{cnt})[1:]

		account := hex.EncodeToString(crc)

		chk := types.Account(account + "00").Checksum()

		id := types.Account(fmt.Sprintf("%v%02d", account, chk))

		info, err := l.AccountInfo(ctx, id)
		if err != nil {
			return types.AllAccounts, err
		}

		if info == nil {
			return id, nil
		}

		cnt++
		if cnt > 10 {
			return "", NewError(InternalError, "cant generate a new id")
		}
	}
}

func (l *Ledger) Health(ctx context.Context) (*schema.DatabaseHealthResponse, error) {
	return l.client.Health(ctx)
}
