package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codenotary/immudb/pkg/api/schema"
	immudb "github.com/codenotary/immudb/pkg/client"
)

type Client struct {
	client   immudb.ImmuClient
	options  *immudb.Options
	limit    uint32
	verified bool
}

func New(ctx context.Context, user string, password string, db string, options ...ClientOption) (*Client, error) {
	cl := &Client{}

	for _, option := range options {
		if option != nil {
			option.Set(cl)
		}
	}

	client, err := immudb.NewImmuClient(cl.options)
	if err != nil {
		return nil, err
	}

	err = client.OpenSession(ctx, []byte(user), []byte(password), db)
	if err != nil {
		return nil, err
	}

	cl.client = client

	return cl, nil
}

func (c *Client) Close(ctx context.Context) {
	c.client.CloseSession(ctx)
}

func (c *Client) DatabaseExist(ctx context.Context, name string) (bool, error) {
	resp, err := c.client.DatabaseListV2(ctx)
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
	return err
}

func (c *Client) UnloadDatabase(ctx context.Context, name string) error {
	_, err := c.client.UnloadDatabase(ctx, &schema.UnloadDatabaseRequest{
		Database: name,
	})

	return err
}

func (c *Client) DeleteDatabase(ctx context.Context, name string) error {
	_, err := c.client.DeleteDatabase(ctx, &schema.DeleteDatabaseRequest{
		Database: name,
	})

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
	if err != nil {
		return 0, err
	}

	return tx.Id, nil
}

func (c *Client) set(ctx context.Context, key []byte, value []byte) (*schema.TxHeader, error) {
	if c.verified {
		return c.client.VerifiedSet(ctx, key, value)
	} else {
		return c.client.Set(ctx, key, value)
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

		tx, err := c.client.VerifiedSet(ctx, []byte(key), data)
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	}
}

func (c *Client) get(ctx context.Context, key []byte) (*schema.Entry, error) {
	if c.verified {
		return c.client.VerifiedGet(ctx, key)
	} else {
		return c.client.Get(ctx, key)
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
		return c.client.VerifiedGetAt(ctx, key, tx)
	} else {
		return c.client.GetAt(ctx, key, tx)
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
	return c.client.TxByID(ctx, id)
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

func (c *Client) ScanSet(ctx context.Context, set string, f func(context.Context, *schema.Tx) (bool, error)) error {
	req := &schema.ZScanRequest{
		Set: []byte(set),
	}

	list, err := c.client.ZScan(ctx, req)
	if err != nil {
		return err
	}

	for _, ze := range list.Entries {
		_ = ze
	}

	return nil
}

func (c *Client) Health(ctx context.Context) (*schema.DatabaseHealthResponse, error) {
	return c.client.Health(ctx)
}
