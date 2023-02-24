package cluster

import (
	"29go_redis/resp/client"
	"context"
	"errors"
	pool "github.com/jolestar/go-commons-pool/v2"
)

type connectionFactory struct {
	Peer string
}

func (f connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	//panic("implement me")
	c, err := client.MakeClient(f.Peer)
	if err != nil {
		return nil, err
	}
	c.Start()
	return pool.NewPooledObject(c), nil

}

func (f connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	//panic("implement me")
	c, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("type mismatch")
	}
	c.Close()
	return nil
}

func (f connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	//panic("implement me")
	return true
}

func (f connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	//panic("implement me")
	return nil
}

func (f connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	//panic("implement me")
	return nil
}
