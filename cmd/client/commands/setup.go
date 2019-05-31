package commands

import (
	"fmt"
	"strings"

	"github.com/dappledger/AnnChain/cmd/client/cluster/docker"
	"github.com/dappledger/AnnChain/cmd/client/cluster/kubernets"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"gopkg.in/urfave/cli.v1"
)

var (
	SetupCommand = cli.Command{
		Name:  "setup",
		Usage: "generate docker-compose.yml to run chain and block-browser",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "crypto_type",
				Value: "ZA",
				Usage: "ZA default is ZA",
			},
			cli.IntFlag{
				Name:  "num",
				Value: 1,
				Usage: "number of validators",
			},
			cli.StringFlag{
				Name:  "rendez_image",
				Value: "",
				Usage: "",
			},
			cli.StringFlag{
				Name:  "block_browser_image",
				Value: "",
				Usage: "",
			},
		},
		Subcommands: cli.Commands{
			{
				Name:   "docker-compose",
				Action: SetupDockerCompose,
			}, {
				Name:   "k8s",
				Action: SetupK8s,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name_prefix",
						Value: "",
						Usage: "name prefix for pod or svc name",
					},
					cli.StringFlag{
						Name:  "namespace",
						Value: "default",
						Usage: "k8s namespace",
					},
					cli.StringFlag{
						Name:  "chain_id",
						Value: "9012",
						Usage: "chain id",
					},
					cli.StringFlag{
						Name:  "output_dir",
						Value: "",
						Usage: "output dir for charts",
					},
				},
			},
		},
	}
)

//func defaultImage(ctx *cli.Context) (r string, i string, b string) {
func defaultImage(ctx *cli.Context) (r string, b string) {
	suffix := ""

	if r = ctx.Parent().String("rendez_image"); r == "" {
		r = fmt.Sprintf("registry.cn-shanghai.aliyuncs.com/anlink/rendez%s:latest", suffix)
	}

	if b = ctx.Parent().String("block_browser_image"); b == "" {
		b = fmt.Sprintf("registry.cn-shanghai.aliyuncs.com/anlink/block-browser%s:latest", suffix)
	}
	return
}

func SetupDockerCompose(ctx *cli.Context) error {

	num := ctx.Parent().Int("num")
	opt := docker.Option{
		ValidatorNum: num,
		IPPrefix:     "192.168.10",
		HasBrowser:   true,
		HasAPI:       true,
		CryptoType:   ctx.Parent().String("crypto_type"),
	}
	opt.RendezImage, opt.BlockBrowserImage = defaultImage(ctx)
	crypto.NodeInit(opt.CryptoType)

	compose := docker.NewDockerCompose(&opt)
	cfg, err := compose.GenerateConfig()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(cfg)
	return nil
}

func SetupK8s(ctx *cli.Context) error {

	num := ctx.Parent().Int("num")

	parseImage := func(name string) *kubernets.Image {
		ss := strings.Split(name, ":")
		return &kubernets.Image{
			Name: ss[0],
			Tag:  ss[1],
		}
	}

	opt := kubernets.Option{
		ValidatorNum: num,
		HasBrowser:   true,
		HasAPI:       true,
		NamePrefix:   ctx.String("name_prefix"),
		ChainID:      ctx.String("chain_id"),
		CryptoType:   ctx.Parent().String("crypto_type"),
		Namespace:    ctx.String("namespace"),
		OutputDir:    ctx.String("output_dir"),
	}

	if opt.OutputDir == "" {
		return cli.NewExitError("output_dir empty", 1)
	}

	r, b := defaultImage(ctx)
	opt.Rendez, opt.BlockBrowser = parseImage(r), parseImage(b)
	crypto.NodeInit(opt.CryptoType)

	if _, err := kubernets.NewHelm(opt); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}
