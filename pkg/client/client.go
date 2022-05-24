package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codenotary/immudb/pkg/api/schema"
	immudb "github.com/codenotary/immudb/pkg/client"
)

type Client struct {
	client  immudb.ImmuClient
	options *immudb.Options
	limit   uint32
}

func New(ctx context.Context, user string, password string, db string, options ...ClientOption) (*Client, error) {
	cl := &Client{}

	for _, option := range options {
		option.Set(cl)
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

func (c *Client) Set(ctx context.Context, key string, value interface{}) (uint64, error) {
	switch v := value.(type) {
	case string:
		tx, err := c.client.Set(ctx, []byte(key), []byte(v))
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	case []byte:
		tx, err := c.client.Set(ctx, []byte(key), v)
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return 0, nil
		}

		tx, err := c.client.Set(ctx, []byte(key), data)
		if err != nil {
			return 0, err
		}

		return tx.Id, nil
	}
}

func (c *Client) Get(ctx context.Context, key string) (*schema.Entry, error) {
	tx, err := c.client.Get(ctx, []byte(key))
	if err != nil {
		return nil, fmt.Errorf("cant get key %v: %v", key, err)
	}

	return tx, nil
}

func (c *Client) GetAt(ctx context.Context, key string, txID uint64) (*schema.Entry, error) {
	tx, err := c.client.GetAt(ctx, []byte(key), txID)
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

/*
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
*/

func (c *Client) ScanAll(ctx context.Context, prefix string, desc bool, f func(context.Context, int, *schema.Entry) (bool, error)) error {
	last := []byte(nil)
	running := true

	for running {
		scanReq := &schema.ScanRequest{
			Prefix:  []byte(prefix),
			Limit:   uint64(c.limit),
			SeekKey: last,
			Desc:    desc,
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

func (c *Client) AllTx(ctx context.Context, desc bool, f func(context.Context, *schema.Tx) (bool, error)) error {
	last := uint64(2)
	running := true

	for running {
		scanReq := &schema.TxScanRequest{
			InitialTx: last,
			Limit:     c.limit,
			Desc:      desc,
		}

		list, err := c.client.TxScan(ctx, scanReq)
		if err != nil {
			return fmt.Errorf("error tx scan: %v", err)
		}

		for _, tx := range list.Txs {
			ok, err := f(ctx, tx)
			if err != nil {
				return err
			}

			if !ok {
				running = false
				break
			}
		}

		if !running || len(list.Txs) < int(scanReq.Limit) {
			break
		}

		last = list.Txs[len(list.Txs)-1].Header.Id
	}

	return nil
}
