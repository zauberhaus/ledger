package ledger

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"

	"github.com/ec-systems/core.ledger.tool/pkg/client"
	"github.com/ec-systems/core.ledger.tool/pkg/ledger/index"
	"github.com/ec-systems/core.ledger.tool/pkg/logger"
	"github.com/ec-systems/core.ledger.tool/pkg/types"
)

const (
	IDLength = 6
)

type Ledger struct {
	client   *client.Client
	overdraw bool
	multi    bool

	assets   types.Assets
	statuses types.Statuses
}

func New(client *client.Client, options ...LedgerOption) *Ledger {
	ledger := &Ledger{
		client: client,
	}

	for _, option := range options {
		option.Set(ledger)
	}

	return ledger
}

func (l *Ledger) Assets() types.Assets {
	return l.assets
}

func (l *Ledger) Statuses() types.Statuses {
	return l.statuses
}

func (l *Ledger) AccountInfo(ctx context.Context, account types.Account) (*types.AccountInfo, error) {
	if account == "" {
		return nil, fmt.Errorf("account is mandatory")
	}

	var info *types.AccountInfo

	err := l.ForEach(ctx, string(index.Account.Key(account)), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		info = &types.AccountInfo{
			Account:  tx.Account,
			Customer: tx.Customer,
			Asset:    tx.Asset,
		}

		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf("get account info failed: %v", err)
	}

	return info, nil
}

func (l *Ledger) Accounts(ctx context.Context, customer string, asset types.Asset) ([]types.Account, error) {
	accounts := []types.Account{}

	if customer == "" {
		return nil, fmt.Errorf("accounts: customer is mandatory")
	}

	err := l.ForEach(ctx, index.Customer.Accounts(customer, asset), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		if customer == tx.Customer {
			accounts = append(accounts, tx.Account)
		} else {
			logger.Errorf("accounts: unexpected customer %v!=%v in %v", customer, tx.Customer, tx.tx)
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("list accounts failed: %v", err)
	}

	return accounts, nil
}

func (l *Ledger) Customers(ctx context.Context) ([][]string, error) {

	customers := [][]string{}

	err := l.ForEach(ctx, index.Customer.All(), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		customers = append(customers, []string{tx.Customer, tx.Account.String(), tx.Asset.String()})
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("list accounts failed: %v", err)
	}

	return customers, nil
}

func (l *Ledger) Balance(ctx context.Context, customer string, asset types.Asset, account types.Account, status types.Status) (map[types.Asset]*types.Balance, error) {
	assets := map[types.Asset]*types.Balance{}

	if customer == "" {
		return nil, fmt.Errorf("balance: customer is mandatory")
	}

	err := l.ForEach(ctx, index.Transaction.Scan(customer, asset, account), false, func(ctx context.Context, tx *Transaction) (bool, error) {

		if customer == tx.Customer {
			balance, ok := assets[tx.Asset]
			if !ok {
				balance = types.NewBalance(status)
				assets[tx.Asset] = balance
			}

			balance.Add(tx.Account, tx.Amount, tx.Status)
		} else {
			logger.Errorf("balance: unexpected customer %v!=%v in %v", customer, tx.Customer, tx.tx)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return assets, nil
}

func (l *Ledger) Transactions(ctx context.Context, customer string, asset types.Asset, account types.Account, f func(context.Context, *Transaction) (bool, error)) error {
	return l.ForEach(ctx, index.Transaction.Scan(customer, asset, account), false, func(ctx context.Context, tx *Transaction) (bool, error) {
		if customer != "" && customer != tx.Customer {
			return false, fmt.Errorf("invalid customer %v in tx %v (%v)", tx.Customer, tx.ID, customer)
		}

		if asset != types.AllAssets && asset != tx.Asset {
			return false, fmt.Errorf("invalid asset %v in tx %v (%v)", tx.Asset, tx.ID, asset)
		}

		if account != types.AllAccounts && account != tx.Account {
			return false, fmt.Errorf("invalid account %v in tx %v (%v)", tx.Account, tx.ID, account)
		}

		return f(ctx, tx)
	})
}

func (l *Ledger) Get(ctx context.Context, transaction string) (*Transaction, error) {
	entry, err := l.client.Get(ctx, string(index.Key.Key(transaction)))
	if err != nil {
		return nil, err
	}

	tx, err := ParseTransaction(entry.Tx, string(entry.Key), entry.Value)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (l *Ledger) Status(ctx context.Context, key string, status types.Status) (*Transaction, error) {
	tx, err := l.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if tx.Status != status {
		tx.Status = status

		txID, _, err := l.Create(ctx, tx)
		if err != nil {
			return nil, err
		}

		tx.tx = txID
	}

	return tx, nil
}

func (l *Ledger) CreateTx(ctx context.Context, customer string, asset types.Asset, amount float64, options ...TransactionOption) (*Transaction, error) {
	if amount == 0 {
		return nil, fmt.Errorf("transaction for customer %v with 0 %v", customer, asset)
	}

	id, err := l.NewID()
	if err != nil {
		return nil, err
	}

	tx := &Transaction{
		Date:     time.Now(),
		ID:       id,
		Customer: customer,

		Status:  types.Created,
		Asset:   asset,
		Amount:  amount,
		Version: uint16(Version),
	}

	for _, option := range options {
		option.Set(tx)
	}

	if tx.Account == "" {
		accounts, err := l.Accounts(ctx, customer, asset)
		if err != nil {
			return nil, err
		}

		if len(accounts) == 0 {
			account, err := types.NewAccount(customer, asset)
			if err != nil {
				return nil, err
			}

			tx.Account = account
		} else if len(accounts) == 1 {
			tx.Account = accounts[0]
		} else if tx.Amount >= 0 {
			return nil, fmt.Errorf("more than one account found for customer %v", customer)
		}
	} else {
		info, err := l.AccountInfo(ctx, tx.Account)
		if err != nil {
			return nil, fmt.Errorf("failed to read account %v info", tx.Account)
		}

		if !l.multi && info == nil {
			return nil, fmt.Errorf("account %v not found", tx.Account)

		}

		if info != nil {
			if info.Customer != tx.Customer {
				return nil, fmt.Errorf("invalid customer %v for account %v (%v)", tx.Customer, tx.Account, info.Customer)
			}

			if info.Asset != tx.Asset {
				return nil, fmt.Errorf("invalid asset %v for account %v (%v)", tx.Asset, tx.Account, info.Asset)
			}
		}
	}

	if tx.Amount < 0 && !l.overdraw {
		balances, err := l.Balance(ctx, customer, asset, tx.Account, types.Created)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer %v balance: %v", customer, err)
		}

		balance, ok := balances[asset]
		if !ok {
			return nil, fmt.Errorf("no %v account found for customer %v", asset, customer)
		}

		if balance.Sum+amount < 0 {
			return nil, fmt.Errorf("balance too low to remove %v %f for customer %v", asset, -amount, customer)
		}

		for k, v := range balance.Accounts {
			if v.Sum+amount >= 0 {
				tx.Account = k
				break
			}
		}

		if tx.Account == "" {
			return nil, fmt.Errorf("no account found with enough balance to remove %v %f for customer %v", asset, -amount, customer)
		}
	}

	if ok := tx.Account.Check(); !ok {
		return nil, fmt.Errorf("invalid checksum for account %v", tx.Account)
	}

	txID, key, err := l.Create(ctx, tx)
	if err != nil {
		return nil, err
	}

	tx.tx = txID
	tx.key = key

	return tx, nil
}

func (l *Ledger) Create(ctx context.Context, tx *Transaction) (uint64, string, error) {
	tx.Date = time.Now()

	if tx.Created.IsZero() {
		tx.Created = tx.Date
	}

	if !tx.Account.Check() {
		return 0, "", fmt.Errorf("checksum check failed for '%v'", tx.Account)
	}

	if !tx.Asset.Check(l.assets) {
		return 0, "", fmt.Errorf("invalid asset '%v'", tx.Asset)
	}

	if tx.Customer == "" {
		return 0, "", fmt.Errorf("customer is empty")
	}

	if tx.ID == "" {
		return 0, "", fmt.Errorf("customer '%v' transaction id is empty", tx.Customer)
	}

	if tx.Amount == 0 {
		return 0, "", nil
	}

	kv := &schema.Op_Kv{
		Kv: &schema.KeyValue{
			Key:   index.Key.Key(tx.ID),
			Value: tx.Bytes(),
		},
	}

	var order *schema.Op_Ref
	if tx.Order != "" || tx.Item != "" {
		order = &schema.Op_Ref{
			Ref: &schema.ReferenceRequest{
				ReferencedKey: kv.Kv.Key,
				Key:           index.Order.Key(tx.Order, tx.Item, tx.ID),
				BoundRef:      false,
			},
		}
	}

	transaction := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Transaction.Key(tx.Customer, tx.Asset, tx.Account, tx.ID),
			BoundRef:      false,
		},
	}

	customer := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Customer.Key(tx.Customer, tx.Asset, tx.Account),
			BoundRef:      false,
		},
	}

	account := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Account.Key(tx.Account),
			BoundRef:      false,
		},
	}

	asset := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Asset.Key(tx.Asset, tx.Customer, tx.Account),
			BoundRef:      false,
		},
	}

	/*
		zasset := &schema.Op_ZAdd{
			ZAdd: &schema.ZAddRequest{
				Set: []byte(tx.Asset),
				Key: kv.Kv.Key,
			},
		}
	*/

	txID, err := l.client.Exec(ctx, kv, order, transaction, customer, account, asset)
	return txID, string(kv.Kv.Key), err
}

func (l *Ledger) Add(ctx context.Context, customer string, asset types.Asset, amount float64, options ...TransactionOption) (*Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("can't add %v %v", asset, amount)
	}

	return l.CreateTx(ctx, customer, asset, amount, options...)
}

func (l *Ledger) Remove(ctx context.Context, customer string, asset types.Asset, amount float64, options ...TransactionOption) (*Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("can't remove %v %v", asset, amount)
	}

	return l.CreateTx(ctx, customer, asset, amount*-1, options...)
}

func (l *Ledger) History(ctx context.Context, id string, f func(ctx context.Context, tx *Transaction) (bool, error)) error {
	return l.client.History(ctx, index.Key.ID(id), func(ctx context.Context, e *schema.Entry) (bool, error) {
		if e.Value[0] == 0 {
			e, err := l.client.GetAt(ctx, string(e.Key), e.Tx)
			if err != nil {
				return true, fmt.Errorf("failed to read the transaction (%v): %v", e.Tx, string(e.Value))
			}

			tx, err := ParseTransaction(e.Tx, string(e.Key), e.Value)
			if err != nil {
				return true, fmt.Errorf("failed to parse the transaction (%v): %v", string(e.Value), err)
			}

			return f(ctx, tx)
		} else {
			tx, err := ParseTransaction(e.Tx, string(e.Key), e.Value)
			if err != nil {
				return true, fmt.Errorf("failed to parse the transaction (%v): %v", string(e.Value), err)
			}

			return f(ctx, tx)
		}
	})
}

func (l *Ledger) ForEach(ctx context.Context, prefix string, desc bool, f func(context.Context, *Transaction) (bool, error)) error {
	return l.client.ScanAll(ctx, prefix, false, func(ctx context.Context, i int, e *schema.Entry) (bool, error) {
		tx, err := ParseTransaction(e.Tx, string(e.Key), e.Value)
		if err != nil {
			return true, fmt.Errorf("failed to parse the transaction (%v): %v", err, string(e.Value))
		}

		return f(ctx, tx)
	})
}

func (l *Ledger) NewID() (string, error) {
	bytes := make([]byte, IDLength)

	err := binary.Read(rand.Reader, binary.BigEndian, &bytes)
	if err != nil {
		return "", fmt.Errorf("error generate random id: %v", err)
	}

	return hex.EncodeToString(bytes), nil
}

func (l *Ledger) ParseAsset(text string) (types.Asset, error) {
	return l.assets.Parse(text)
}
