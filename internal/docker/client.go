package docker

import (
	"context"

	"github.com/docker/docker/client"
)

var cli *client.Client

func GetClient() (*client.Client, error) {
	if cli != nil {
		return cli, nil
	}

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func Ping(ctx context.Context) error {
	c, err := GetClient()
	if err != nil {
		return err
	}

	_, err = c.Ping(ctx)
	return err
}
