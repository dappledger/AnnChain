# What is angine?

Angine is a completely self-contained blockchain consensus engine. 
At the core, we use Tendermint BFT. It is an implementation of a Byzantine Fault Tolerant PoS consensus algorithm. So thank you Tendermint team.
We just wrap those many things that tendermint already offered together under a concept, "Angine". 



# Structure of Angine

```shell
├── angine.go
├── blockchain
│   ├── pool.go
│   ├── pool_test.go
│   ├── reactor.go
│   └── store.go
├── config
│   ├── config.go
│   └── templates.go
├── consensus
│   ├── byzantine_test.go
│   ├── common.go
│   ├── common_test.go
│   ├── height_vote_set.go
│   ├── height_vote_set_test.go
│   ├── mempool_test.go
│   ├── reactor.go
│   ├── reactor_test.go
│   ├── README.md
│   ├── replay.go
│   ├── replay_test.go
│   ├── state.go
│   ├── state_test.go
│   ├── test_data
│   │   ├── build.sh
│   │   ├── empty_block.cswal
│   │   ├── README.md
│   │   ├── small_block1.cswal
│   │   └── small_block2.cswal
│   ├── ticker.go
│   ├── version.go
│   └── wal.go
├── LICENSE
├── log.go
├── mempool
│   ├── mempool.go
│   ├── mempool_test.go
│   └── reactor.go
├── plugin
│   ├── init.go
│   └── specialop.go
├── README.md
├── refuse_list
│   ├── refuse_list.go
│   └── refuse_list_test.go
├── specialop.go
├── state
│   ├── errors.go
│   ├── execution.go
│   ├── execution_test.go
│   ├── plugin.go
│   ├── state.go
│   ├── state_test.go
│   ├── tps.go
│   └── tps_test.go
├── trace
│   ├── reactor.go
│   └── router.go
└── types
    ├── application.go
    ├── block.go
    ├── block_meta.go
    ├── canonical_json.go
    ├── common.go
    ├── core.go
    ├── events.go
    ├── genesis.go
    ├── hooks.go
    ├── keys.go
    ├── mempool.go
    ├── part_set.go
    ├── part_set_test.go
    ├── priv_validator.go
    ├── proposal.go
    ├── proposal_test.go
    ├── query.go
    ├── resultcode.go
    ├── result.go
    ├── rpc.go
    ├── signable.go
    ├── specialOP.go
    ├── tx.go
    ├── validator.go
    ├── validator_set.go
    ├── validator_set_test.go
    ├── vote.go
    ├── vote_set.go
    ├── vote_set_test.go
    └── vote_test.go
```

This is directory structure of Angine, you can see that we have packed every module under its own directory. This give you a clear view of the parts making up an Angine.

1.  state/ is the running state of the Angine, which is driven by blockchain/, mempool/ and consensus/
2.  blockchain/ takes charge of syncing blocks, loading blocks, persisting blocks and anything that physically related to "block"
3.  mempool/ takes charge of buffering effective transactions and reaching an agreement about the order of transactions
4.  consensus/ takes charge of gossipping between peers, consensus algorithm related data stream
5.  trace/ is just another reactor module used for specialop votes currently




# What we have fulfilled

1.  configurable CA based on asymmetric cyrpto

2.  Dynamically changing ValidatorSet of ConsensusState

3.  Two kinds of transactions, normal and special, are totally isolated. Special tx will only be processed by plugins by default.

4.  Angine plugins

5.  Node joining will automatically download genesisfile from the first seed listed in config file




# How to use

### Configuration

For angine, every thing angine needs is located at a universal directory, named "angine_runtime". By default, the directory sits at `$HOME/.angine`. And we also provide an environment variable to override the default: `ANGINE_RUNTIME`.

The auto generated runtime looks like this:

```shell
├── config.toml
├── data
├── genesis.json
└── priv_validator.json
```

* config.toml contains all your custom configs for Angine
* data contains all the data generated when Angine starts running
* genesis.json defines genesis state of your blockchain
* priv_validator.json contains your node's private/public key pair

Let me just show you a simplified config.toml:

```toml
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

environment = "development"
db_backend = "leveldb"
moniker = "anonymous"
p2p_laddr = "tcp://0.0.0.0:46656"
seeds = ""

# auth by ca general switch
auth_by_ca = true

# whether non-validator nodes need auth by ca, only effective when auth_by_ca is true
non_validator_auth_by_ca = true

# auth signature from CA
signbyCA = ""

fast_sync = true

skip_upnp = true

log_path = ""

#log_level:
        # -1 DebugLevel logs are typically voluminous, and are usually disabled in production.
        #  0 InfoLevel is the default logging priority.
        #  1 WarnLevel logs are more important than Info, but don't need individual human review.
        #  2 ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs.
        #  3 DPanicLevel logs are particularly important errors. In development the logger panics after writing the message.
        #  4 PanicLevel logs a message, then panics.
        #  5 FatalLevel logs a message, then calls os.Exit(1)
```



### Default CA signature

Suppose you want to join a chain with a public key "123456" and its chainID is "abc". The patten is like $publickey$$chainID$:

```shell
123456abc
```

This message above is what you should give the CA node of that chain and expect a corresponding signature which means an authorized identity to join.

Further more, a blockchain can choose to use 3 different auth strategies: 

* auth by default 
* only validators need auth
* no auth at all 

### Initialize Angine

This is how you initialize an angine. 

```go
angine.Initialize(&angine.Tunes{Conf: conf})
```

The "angine.Initialize" will handle the generation of default configs, genesis file and private key. You must only do this once for a particular chain, otherwise, your chain id will be different for sure.

### Construct a Tunes.

```go
type Tunes struct {
    Runtime string
    Conf    *viper.Viper
}
```

This struct contains 2 fields and you only have to fill one:

1.  Runtime is a path that contains all the auto-generated files. So provided this will just generate everything under this path with random chainID and private/pub key pair.

2.  Conf contains any config that you want to override the defaults. Say, you wanna use some cli args to override the default configs, this is the thing you should look into.

    After, you probably need to edit those files mannually to get exactly what you want. Everything in the files are straigt forward.


### New an Angine instance and start it

First, you need to import angine into your project :-) then, 

```go
// this line should be changed to github path accordingly
import "github.com/dappledger/AnnChain/angine" 

...

mainAngine := angine.NewAngine(&angine.Tunes{Conf: conf})

...

mainAngine.Start()
```

That is all.

### Manage Validator Set

Angine is designed to be capable of chaning validators dynamically. 

First of all, all new connected nodes are by default non-validator. If the node wants to become a validator, a change_validator command is required. In angine world, such a command is called special operation and change_validator must come from CA node.

Suppose you are a CA node, use the following command to initiate a special operation to change a normal node into a validator:

```shell
anntool --backend=${CA's RPC address} --target="${chainid}" special change_validator --privkey="${CA's privatekey}" --power=${voting power} --isCA="{false/true}" --validator_pubkey="${the node's publickey}"
```

