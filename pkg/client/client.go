package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codenotary/immudb/pkg/api/schema"
	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/ec-systems/core.ledger.server/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client   immudb.ImmuClient
	options  *immudb.Options
	limit    uint32
	verified bool

	user     []byte
	password []byte
	db       string
}

func New(ctx context.Context, user string, password string, db string, options ...ClientOption) (*Client, error) {
	cl := &Client{
		user:     []byte(user),
		password: []byte(password),
		db:       db,
	}

	for _, option := range options {
		if option != nil {
			option.Set(cl)
		}
	}

	client, err := immudb.NewImmuClient(cl.options)
	if err != nil {
		return nil, err
	}

	err = client.OpenSession(ctx, cl.user, cl.password, cl.db)
	if err != nil {
		return nil, err
	}

	logger.Infof("Connected to immudb database '%v' (%v:%v)", cl.db, cl.options.Address, cl.options.Port)

	cl.client = client

	return cl, nil
}

func (c *Client) Close(ctx context.Context) error {
	err := c.client.CloseSession(ctx)
	if err == nil {
		logger.Info("Database disconnected")
	} else {
		logger.Errorf("Database disconnect failed: %v", err)
	}

	return err
}

func (c *Client) DatabaseExist(ctx context.Context, name string) (bool, error) {
	resp, err := c.client.DatabaseListV2(ctx)
	for !c.checkSessionError(ctx, err) {
		resp, err = c.client.DatabaseListV2(ctx)
	}

	if err != nil {
		return false, err
	}

	for _, db := range resp.Databases {
		if db.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) CreateDatabase(ctx context.Context, name string) error {
	_, err := c.client.CreateDatabaseV2(ctx, name, nil)
	for !c.checkSessionError(ctx, err) {
		_, err = c.client.CreateDatabaseV2(ctx, name, nil)
	}

	return err
}

func (c *Client) UnloadDatabase(ctx context.Context, name string) error {
	_, err := c.client.UnloadDatabase(ctx, &schema.UnloadDatabaseRequest{
		Database: name,
	})

	for !c.checkSessionError(ctx, err) {
		_, err = c.client.UnloadDatabase(ctx, &schema.UnloadDatabaseRequest{
			Database: name,
		})
	}

	return err
}

func (c *Client) DeleteDatabase(ctx context.Context, name string) error {
	_, err := c.client.DeleteDatabase(ctx, &schema.DeleteDatabaseRequest{
		Database: name,
	})

	for !c.checkSessionError(ctx, err) {
		_, err = c.client.DeleteDatabase(ctx, &schema.DeleteDatabaseRequest{
			Database: name,
		})
	}

	return err
}

func (c *Client) Exec(ctx context.Context, operations ...interface{}) (uint64, error) {

	ops := []*schema.Op{}
	pre := []*schema.Precondition{}

	for _, op := range operations {
		if op == nil {
			continue
		}

		switch o := op.(type) {
		case *schema.Op_Ref:
			if o != nil {
				ops = append(ops, &schema.Op{
					Operation: o,
				})
			}
		case *schema.Op_Kv:
			if o != nil {
				ops = append(ops, &schema.Op{
					Operation: o,
				})
			}
		case *schema.Op_ZAdd:
			if o != nil {
				ops = append(ops, &schema.Op{
					Operation: o,
				})
			}
		case *schema.Precondition_KeyMustExist:
			if o != nil {
				pre = append(pre, &schema.Precondition{
					Precondition: o,
				})
			}
		case *schema.Precondition_KeyMustNotExist:
			if o != nil {
				pre = append(pre, &schema.Precondition{
					Precondition: o,
				})
			}
		case *schema.Precondition_KeyNotModifiedAfterTX:
			if o != nil {
				pre = append(pre, &schema.Precondition{
					Precondition: o,
				})
			}
		}
	}

	tx, err := c.client.ExecAll(ctx, &schema.ExecAllRequest{
		Operations:    ops,
		Preconditions: pre,
	})

	for !c.checkSessionError(ctx, err) {
		tx, err = c.client.ExecAll(ctx, &schema.ExecAllRequest{
			Operations:    ops,
			Preconditions: pre,
		})
	}

	if err != nil {
		return 0, err
	}

	return tx.Id, nil
}

func (c *Client) set(ctx context.Context, key []byte, value []byte) (*schema.TxHeader, error) {
	if c.verified {
		header, err := c.client.VerifiedSet(ctx, key, value)
		for !c.checkSessionError(ctx, err) {
			header, err = c.client.VerifiedSet(ctx, key, value)
		}
		return header, err
	} else {
		header, err := c.client.Set(ctx, key, value)
		for !c.checkSessionError(ctx, err) {
			header, err = c.client.Set(ctx, key, value)
		}
		return header, err
	}
}

func (c *Client) Set(ctx context.Context, key []byte, value interface{}) (uint64, error) {

	switch v := value.(type) {
	case string:
		tx, err := c.set(ctx, []byte(key), []byte(v))
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	case []byte:
		tx, err := c.set(ctx, []byte(key), v)
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return 0, nil
		}

		tx, err := c.set(ctx, []byte(key), data)
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	}
}

func (c *Client) get(ctx context.Context, key []byte) (*schema.Entry, error) {
	if c.verified {
		entry, err := c.client.VerifiedGet(ctx, key)
		for !c.checkSessionError(ctx, err) {
			entry, err = c.client.VerifiedGet(ctx, key)
		}
		return entry, err
	} else {
		entry, err := c.client.Get(ctx, key)
		for !c.checkSessionError(ctx, err) {
			entry, err = c.client.Get(ctx, key)
		}
		return entry, err
	}
}

func (c *Client) Get(ctx context.Context, key string) (*schema.Entry, error) {
	tx, err := c.get(ctx, []byte(key))
	if err != nil {
		return nil, fmt.Errorf("cant get key %v: %v", key, err)
	}

	return tx, nil
}

func (c *Client) getAt(ctx context.Context, key []byte, tx uint64) (*schema.Entry, error) {
	if c.verified {
		entry, err := c.client.VerifiedGetAt(ctx, key, tx)
		for !c.checkSessionError(ctx, err) {
			entry, err = c.client.VerifiedGetAt(ctx, key, tx)
		}
		return entry, err
	} else {
		entry, err := c.client.GetAt(ctx, key, tx)
		for !c.checkSessionError(ctx, err) {
			entry, err = c.client.GetAt(ctx, key, tx)
		}
		return entry, err
	}
}

func (c *Client) GetAt(ctx context.Context, key string, txID uint64) (*schema.Entry, error) {
	tx, err := c.getAt(ctx, []byte(key), txID)
	if err != nil {
		return nil, fmt.Errorf("cant get key %v: %v", key, err)
	}

	return tx, nil
}

func (c *Client) GetTx(ctx context.Context, id uint64) (*schema.Tx, error) {
	tx, err := c.client.TxByID(ctx, id)
	for !c.checkSessionError(ctx, err) {
		tx, err = c.client.TxByID(ctx, id)
	}
	return tx, err
}

func (c *Client) History(ctx context.Context, key string, f func(context.Context, *schema.Entry) (bool, error)) error {
	offset := uint64(0)
	running := true

	for running {

		req := &schema.HistoryRequest{
			Key:    []byte(key),
			Limit:  int32(c.limit),
			Offset: offset,
		}

		list, err := c.client.History(ctx, req)
		for !c.checkSessionError(ctx, err) {
			list, err = c.client.History(ctx, req)
		}

		if err != nil {
			return fmt.Errorf("failed to read history of key %v: %v", key, err)
		}

		for _, v := range list.Entries {
			ok, err := f(ctx, v)
			if err != nil {
				return err
			}

			if !ok {
				running = false
				break
			}

			offset++
		}

		if !running || len(list.Entries) < int(req.Limit) {
			break
		}
	}

	return nil
}

func (c *Client) LastTX(ctx context.Context) (uint64, error) {
	state, err := c.client.CurrentState(ctx)
	for !c.checkSessionError(ctx, err) {
		state, err = c.client.CurrentState(ctx)
	}

	if err != nil {
		return 0, err
	}

	return state.TxId, nil
}

func (c *Client) Scan(ctx context.Context, prefix string, limit uint64, desc bool) ([]*schema.Entry, error) {

	scanReq := &schema.ScanRequest{
		Prefix: []byte(prefix),
		Limit:  limit,
		Desc:   desc,
	}

	list, err := c.client.StreamScan(ctx, scanReq)
	for !c.checkSessionError(ctx, err) {
		list, err = c.client.StreamScan(ctx, scanReq)
	}

	if err != nil {
		return nil, fmt.Errorf("error scan %v: %v", prefix, err)
	}

	return list.Entries, nil
}

func (c *Client) ScanAll(ctx context.Context, prefix string, desc bool, f func(context.Context, int, *schema.Entry) (bool, error)) error {
	return c.ScanSince(ctx, prefix, desc, 0, f)
}

func (c *Client) ScanSince(ctx context.Context, prefix string, desc bool, since uint64, f func(context.Context, int, *schema.Entry) (bool, error)) error {
	last := []byte(nil)
	running := true

	for running {
		scanReq := &schema.ScanRequest{
			Prefix:  []byte(prefix),
			Limit:   uint64(c.limit),
			SeekKey: last,
			Desc:    desc,
			SinceTx: since,
		}

		list, err := c.client.Scan(ctx, scanReq)
		for !c.checkSessionError(ctx, err) {
			list, err = c.client.Scan(ctx, scanReq)
		}

		if err != nil {
			return fmt.Errorf("error scan %v: %v", prefix, err)
		}

		for i, v := range list.Entries {
			ok, err := f(ctx, i, v)
			if err != nil {
				return err
			}

			if !ok {
				running = false
				break
			}

			if v.ReferencedBy != nil {
				last = v.ReferencedBy.Key
			}
		}

		if !running || last == nil || len(list.Entries) < int(scanReq.Limit) {
			break
		}
	}

	return nil
}

func (c *Client) ScanSet(ctx context.Context, set string, desc bool, f func(context.Context, *schema.ZEntry) (bool, error)) error {
	var last *schema.ZEntry

	running := true

	for running {
		scanReq := &schema.ZScanRequest{
			Set:   []byte(set),
			Limit: uint64(c.limit),
			Desc:  desc,
		}

		if last != nil {
			scanReq.SeekKey = last.Key
			scanReq.SeekScore = last.Score
			scanReq.SeekAtTx = last.AtTx
		}

		list, err := c.client.ZScan(ctx, scanReq)
		for !c.checkSessionError(ctx, err) {
			list, err = c.client.ZScan(ctx, scanReq)
		}

		if err != nil {
			return fmt.Errorf("error scan set %v: %v", set, err)
		}

		for _, v := range list.Entries {
			ok, err := f(ctx, v)
			if err != nil {
				return err
			}

			if !ok {
				running = false
				break
			}
		}

		if !running || len(list.Entries) == 0 || len(list.Entries) < int(scanReq.Limit) {
			break
		}

		last = list.Entries[len(list.Entries)-1]
	}

	return nil
}

func (c *Client) Export(ctx context.Context, tx uint64) (schema.ImmuService_ExportTxClient, error) {
	req := &schema.ExportTxRequest{
		Tx: tx,
	}

	client, err := c.client.ExportTx(ctx, req)
	for !c.checkSessionError(ctx, err) {
		client, err = c.client.ExportTx(ctx, req)
	}

	return client, err
}

func (c *Client) Replicate(ctx context.Context) (schema.ImmuService_ReplicateTxClient, error) {
	client, err := c.client.ReplicateTx(ctx)
	for !c.checkSessionError(ctx, err) {
		client, err = c.client.ReplicateTx(ctx)
	}

	return client, err
}

func (c *Client) Health(ctx context.Context) (*schema.DatabaseHealthResponse, error) {
	response, err := c.client.Health(ctx)
	for !c.checkSessionError(ctx, err) {
		response, err = c.client.Health(ctx)
	}

	return response, err
}

func (c *Client) checkSessionError(ctx context.Context, err error) bool {
	if err == nil {
		return true
	}

	code, ok := status.FromError(err)
	if ok {

		if code.Code() == codes.PermissionDenied {

			c.client.CloseSession(ctx)

			err = c.client.OpenSession(ctx, c.user, c.password, c.db)
			if err != nil {
				logger.Error(err)
				return false
			}

			return false
		}
	}

	return true
}
