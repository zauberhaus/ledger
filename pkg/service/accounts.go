package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/maps"
)

type AccountsService struct {
	chi.Router
	ledger *ledger.Ledger
}

func NewAccountsService(ledger *ledger.Ledger) chi.Router {
	router := chi.NewRouter()
	svc := &AccountsService{
		Router: router,
		ledger: ledger,
	}

	// add new asset
	router.Put("/{holder}/{asset}/{amount}", svc.add)
	// remove asset
	router.Delete("/{holder}/{asset}/{amount}", svc.remove)
	// list holders
	router.Get("/", svc.holders)
	// list accounts with balance
	router.Get("/{holder}", svc.allAccounts)
	// list accounts with balance
	router.Get("/{holder}/{asset}", svc.accounts)
	// list tx from account
	router.Get("/{holder}/{asset}/{account}", svc.transactions)
	// show tx history
	router.Get("/{holder}/{asset}/{account}/{id}", svc.history)
	// set tx status
	router.Patch("/{holder}/{asset}/{account}/{id}/{status}", svc.change)
	// revert a transaction
	router.Delete("/{holder}/{asset}/{account}/{id}", svc.cancel)

	return svc
}

// @Summary      Add Assets
// @Description  Add assets to the ledger
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Param        amount   	path      	string  true  	"Amount"
// @Param        order   	query     	string  false  	"Order ID"
// @Param        item   	query    	string  false  	"Order Item ID"
// @Param        ref   		query      	string 	false	"Reference"
// @Success      200  {object}  service.Transaction
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset}/{amount} [put]
func (a *AccountsService) add(w http.ResponseWriter, r *http.Request) {
	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "empty holder", http.StatusBadRequest)
		return
	}

	amount, err := a.amount(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := a.account(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := a.order(w, r)
	item := a.item(w, r)
	ref := a.reference(w, r)

	tx, err := a.ledger.Add(r.Context(), holder, asset, amount, account, order, item, ref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output := &Transaction{}
	output.Set(a.ledger, tx)
	render.JSON(w, r, output)
}

// @Summary      Remove Assets
// @Description  Remove assets to the ledger
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Param        amount   	path      	string  true  	"Amount"
// @Param        order   	query     	string  false  	"Order ID"
// @Param        item   	query    	string  false  	"Order Item ID"
// @Param        ref   		query      	string 	false	"Reference"
// @Success      200  {object}  service.Transaction
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset}/{amount} [delete]
func (a *AccountsService) remove(w http.ResponseWriter, r *http.Request) {
	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "empty holder", http.StatusBadRequest)
		return
	}

	amount, err := a.amount(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := a.account(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := a.order(w, r)
	item := a.item(w, r)
	ref := a.reference(w, r)

	tx, err := a.ledger.Remove(r.Context(), holder, asset, amount, account, order, item, ref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output := &Transaction{}
	output.Set(a.ledger, tx)
	render.JSON(w, r, output)
}

// @Summary      Revert a Transaction
// @Description  Remove or add assets from ledger by reverting a transaction
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Param        account   	path      	string  true  	"Account"
// @Param        id   		path      	string  true  	"Transaction ID"
// @Success      200  {object}  service.Transaction
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset}/{account}/{id} [delete]
func (a *AccountsService) cancel(w http.ResponseWriter, r *http.Request) {

	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "empty holder", http.StatusBadRequest)
		return
	}

	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if asset == types.AllAssets {
		http.Error(w, "asset is mandatory", http.StatusBadRequest)
		return
	}

	account, err := a.account(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if account == nil {
		http.Error(w, "account is mandatory", http.StatusBadRequest)
		return
	}

	id, err := a.id(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if id.IsEmpty() {
		http.Error(w, "transaction id is mandatory", http.StatusBadRequest)
		return
	}

	in := &ledger.Transaction{
		ID:     id,
		Holder: holder,
		Asset:  asset,
	}

	if account != nil {
		account.Set(in)
	}

	tx, err := a.ledger.Cancel(r.Context(), in.Holder, in.Asset, in.Account, in.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output := &Transaction{}
	output.Set(a.ledger, tx)
	render.JSON(w, r, output)
}

// @Summary      List Holders
// @Description  List all holders in the ledger
// @Tags         Accounts
// @Produce      json
// @Success 	 200 		{array} service.Holder
// @Failure      404
// @Failure      500
// @Router       /accounts/ [get]
func (a *AccountsService) holders(w http.ResponseWriter, r *http.Request) {
	holders := map[string]*Holder{}
	err := a.ledger.Holders(r.Context(), func(holder string, account types.Account, asset types.Asset) (bool, error) {
		h, ok := holders[holder]
		if !ok {
			h = &Holder{
				Name:     holder,
				Accounts: []*Account{},
			}

			holders[holder] = h
		}

		h.Accounts = append(h.Accounts, &Account{
			Account: account.String(),
			Asset:   asset.String(),
		})

		return true, nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(holders) == 0 {
		http.Error(w, "Ledger is empty", http.StatusNotFound)
		return
	}

	result := maps.Values(holders)
	render.JSON(w, r, result)
}

// @Summary      List User Accounts
// @Description  List accounts and balances of a holder
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Success 	 200 		{array} service.Balance
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder} [get]
func (a *AccountsService) allAccounts(w http.ResponseWriter, r *http.Request) {
	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "empty holder", http.StatusBadRequest)
		return
	}

	balances, err := a.ledger.Balance(r.Context(), holder, types.AllAssets, types.AllAccounts, types.AllStatuses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(balances) == 0 {
		http.Error(w, fmt.Sprintf("No accounts for holder %v found\n", holder), http.StatusNotFound)
		return
	}

	list := []*Balance{}

	for k, v := range balances {
		accounts := []*AccountBalance{}

		for a, b := range v.Accounts {
			accounts = append(accounts, &AccountBalance{
				ID:    a.String(),
				Count: b.Count,
				Sum:   b.Sum,
			})
		}

		list = append(list, &Balance{
			Asset:    k.String(),
			Accounts: accounts,
			Count:    v.Count,
			Sum:      v.Sum,
		})
	}

	render.JSON(w, r, list)
}

// @Summary      List Asset Accounts
// @Description  List accounts and balances of a asset of a holder
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  false  	"Asset Symbol"
// @Success 	 200 		{array} service.Balance
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset} [get]
func (a *AccountsService) accounts(w http.ResponseWriter, r *http.Request) {
	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "empty holder", http.StatusBadRequest)
		return
	}

	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balances, err := a.ledger.Balance(r.Context(), holder, asset, types.AllAccounts, types.AllStatuses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(balances) == 0 {
		http.Error(w, fmt.Sprintf("No accounts for holder %v found\n", holder), http.StatusNotFound)
		return
	}

	list := []*Balance{}

	for k, v := range balances {
		accounts := []*AccountBalance{}

		for a, b := range v.Accounts {
			accounts = append(accounts, &AccountBalance{
				ID:    a.String(),
				Count: b.Count,
				Sum:   b.Sum,
			})
		}

		list = append(list, &Balance{
			Asset:    k.String(),
			Accounts: accounts,
			Count:    v.Count,
			Sum:      v.Sum,
		})
	}

	render.JSON(w, r, list)
}

// @Summary      List Transactions
// @Description  List all the transactions of an account
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Param        account   	path      	string  true  	"Account"
// @Success 	 200 		{object} service.Transaction
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset}/{account} [get]
func (a *AccountsService) transactions(w http.ResponseWriter, r *http.Request) {
	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "holder is mandatory", http.StatusBadRequest)
		return
	}

	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if asset == types.AllAssets {
		http.Error(w, "asset is mandatory", http.StatusBadRequest)
		return
	}

	account, err := a.account(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if account == nil {
		http.Error(w, "account is mandatory", http.StatusBadRequest)
		return
	}

	in := &ledger.Transaction{
		Holder: holder,
		Asset:  asset,
	}

	if account != nil {
		account.Set(in)
	}

	txs := []*Transaction{}

	err = a.ledger.Transactions(r.Context(), in.Holder, in.Asset, in.Account, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
		if tx.Holder == in.Holder && tx.Asset == in.Asset && tx.Account == in.Account {
			output := &Transaction{}
			output.Set(a.ledger, tx)
			txs = append(txs, output)
			return true, nil
		} else {
			return false, fmt.Errorf("invalid holder/asset/account combination: %v", in.ID)
		}
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, txs)
}

// @Summary      Show History
// @Description  Show the history of a transaction
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Param        account   	path      	string  true  	"Account"
// @Param        id   		path      	string  true  	"Transaction ID"
// @Success 	 200 		{array} service.Transaction
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset}/{account}/{id} [get]
func (a *AccountsService) history(w http.ResponseWriter, r *http.Request) {
	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "holder is mandatory", http.StatusBadRequest)
		return
	}

	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if asset == types.AllAssets {
		http.Error(w, "asset is mandatory", http.StatusBadRequest)
		return
	}

	account, err := a.account(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if account == nil {
		http.Error(w, "account is mandatory", http.StatusBadRequest)
		return
	}

	id, err := a.id(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if id.IsEmpty() {
		http.Error(w, "transaction id is mandatory", http.StatusBadRequest)
		return
	}

	in := &ledger.Transaction{
		ID:     id,
		Holder: holder,
		Asset:  asset,
	}

	if account != nil {
		account.Set(in)
	}

	txs := []*Transaction{}

	err = a.ledger.History(r.Context(), in.ID, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
		if tx.Holder == in.Holder && tx.Asset == in.Asset && tx.Account == in.Account && tx.ID == in.ID {
			output := &Transaction{}
			output.Set(a.ledger, tx)
			txs = append(txs, output)
			return true, nil
		} else {
			return false, fmt.Errorf("invalid holder/asset/account/id for tx history: %v", in.ID)
		}
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, txs)
}

// @Summary      Change the Transaction Status
// @Description  Change the status of a transaction
// @Tags         Accounts
// @Produce      json
// @Param        holder   	path      	string  true  	"Account Holder"
// @Param        asset   	path      	string  true  	"Asset Symbol"
// @Param        account   	path      	string  true  	"Account"
// @Param        id   		path      	string  true  	"Transaction ID"
// @Param        status   	path      	string  true  	"Transaction Status"
// @Success 	 200 		{object} service.Transaction
// @Failure      404
// @Failure      500
// @Router       /accounts/{holder}/{asset}/{account}/{id}/{status} [patch]
func (a *AccountsService) change(w http.ResponseWriter, r *http.Request) {
	holder := a.holder(w, r)
	if holder == "" {
		http.Error(w, "holder is mandatory", http.StatusBadRequest)
		return
	}

	asset, err := a.asset(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if asset == types.AllAssets {
		http.Error(w, "asset is mandatory", http.StatusBadRequest)
		return
	}

	account, err := a.account(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if account == nil {
		http.Error(w, "account is mandatory", http.StatusBadRequest)
		return
	}

	id, err := a.id(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := a.status(w, r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if status == types.AllStatuses {
		http.Error(w, "status is mandatory", http.StatusBadRequest)
		return
	}

	in := &ledger.Transaction{
		ID:     id,
		Holder: holder,
		Asset:  asset,
		Status: status,
	}

	if account != nil {
		account.Set(in)
	}

	tx, err := a.ledger.Status(r.Context(), in, in.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tx == nil {
		http.Error(w, "same status", http.StatusNotModified)
		return
	}

	output := &Transaction{}
	output.Set(a.ledger, tx)
	render.JSON(w, r, output)
}

func (t *AccountsService) asset(w http.ResponseWriter, r *http.Request) (types.Asset, error) {
	assetID := chi.URLParam(r, "asset")

	if assetID == "" {
		return types.AllAssets, nil
	}

	return t.ledger.SupportedAssets().Parse(assetID)
}

func (t *AccountsService) status(w http.ResponseWriter, r *http.Request) (types.Status, error) {
	statusID := chi.URLParam(r, "status")

	if statusID == "" {
		return types.AllStatuses, nil
	}

	return t.ledger.SupportedStatus().Parse(statusID)
}

func (l *AccountsService) holder(w http.ResponseWriter, r *http.Request) string {
	return chi.URLParam(r, "holder")
}

func (l *AccountsService) amount(w http.ResponseWriter, r *http.Request) (decimal.Decimal, error) {
	amount := chi.URLParam(r, "amount")
	return decimal.NewFromString(amount)
}

func (l *AccountsService) id(w http.ResponseWriter, r *http.Request) (types.ID, error) {
	txid := chi.URLParam(r, "id")
	if txid == "" {
		return types.ZeroID, fmt.Errorf("transaction id is mandatory")
	}

	guid, err := uuid.Parse(txid)
	if err != nil {
		return types.ZeroID, fmt.Errorf("transaction id is invalid: %v", err)
	}

	id := types.ID{UUID: guid}
	if id.IsEmpty() {
		return types.ZeroID, fmt.Errorf("transaction id is empty")
	}

	return id, nil
}

func (l *AccountsService) order(w http.ResponseWriter, r *http.Request) ledger.TransactionOption {
	orderID := r.URL.Query().Get("order")
	if orderID != "" {
		return ledger.OrderID(orderID)
	}

	return nil
}

func (l *AccountsService) item(w http.ResponseWriter, r *http.Request) ledger.TransactionOption {
	orderItemID := r.URL.Query().Get("item")
	if orderItemID != "" {
		return ledger.OrderItemID(orderItemID)
	}

	return nil
}

func (l *AccountsService) reference(w http.ResponseWriter, r *http.Request) ledger.TransactionOption {
	ref := r.URL.Query().Get("ref")
	if ref != "" {
		return ledger.Reference(ref)
	}

	return nil
}

func (l *AccountsService) account(w http.ResponseWriter, r *http.Request) (ledger.TransactionOption, error) {
	accountID := chi.URLParam(r, "account")
	if accountID == "" {
		accountID = r.URL.Query().Get("account")
	}

	if accountID != "" {
		account := types.Account(accountID)
		if !account.Check() {
			return nil, fmt.Errorf("invalid checksum for account %v", account)
		}

		if account != types.AllAccounts {
			return ledger.Account(account), nil
		} else {
			return nil, nil
		}

	}

	return nil, nil
}
