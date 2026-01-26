package main

import (
	"context"
	"fmt"
	"github.com/AtomiCloud/sulfone.boron/docker_executor"
	imageTypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	rt "runtime"
)

func main() {

	app := &cli.App{
		Name: "sulfone-boron",
		Commands: []*cli.Command{
			{
				Name: "s",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
					if err != nil {
						panic(err)
					}
					defer func(dCli *client.Client) {
						_ = dCli.Close()
					}(dCli)
					images, err := dCli.ImageList(ctx, imageTypes.ListOptions{})
					if err != nil {
						panic(err)
					}
					for _, image := range images {
						fmt.Println(image.RepoTags)
					}
					return nil
				},
			},
			{
				Name: "start",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "registry",
						Aliases: []string{"r"},
						Value:   "https://api.zinc.sulfone.raichu.cluster.atomi.cloud",
					},
				},
				Action: func(context *cli.Context) error {
					registry := context.String("registry")
					server(registry)
					return nil
				},
			},
			{
				Name: "setup",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
					if err != nil {
						panic(err)
					}
					defer func(dCli *client.Client) {
						_ = dCli.Close()
					}(dCli)
					cpu := rt.NumCPU()
					d := docker_executor.DockerClient{
						Docker:           dCli,
						Context:          ctx,
						ParallelismLimit: cpu,
					}
					err = d.EnforceNetwork()
					if err != nil {
						fmt.Println("🚨 Error enforcing network", err)
						return err
					}
					fmt.Println("✅ Enforced network")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
