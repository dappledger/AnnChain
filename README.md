<div id="table-of-contents">
<h2>Table of Contents</h2>
<div id="text-table-of-contents">
<ul>
<li><a href="#orgbe773ea">1. What is angine?</a></li>
<li><a href="#orgf0f2108">2. Structure of Angine</a></li>
<li><a href="#org3958c62">3. What we have fulfilled</a></li>
<li><a href="#org5e86273">4. How to use</a>
<ul>
<li><a href="#org9a260de">4.1. Initialize Angine</a>
<ul>
<li><a href="#org19c7657">4.1.1. Construct an AngineTunes.</a></li>
</ul>
</li>
<li><a href="#orgf101d9a">4.2. New an Agnine instance and start it</a></li>
</ul>
</li>
</ul>
</div>
</div>

<a id="orgbe773ea"></a>

# What is angine?

Angine is a completely self-contained blockchain consensus engine. 
At the core, we use Tendermint BFT. It is an implementation of a Byzantine Fault Tolerant PoS consensus algorithm. So thank you Tendermint team.
We just wrap those many things that tendermint already offered together under a concept, "Angine". 


<a id="orgf0f2108"></a>

# Structure of Angine

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
    ├── angine.go
    ├── log.go
    ├── mempool
    │   ├── mempool.go
    │   ├── mempool_test.go
    │   └── reactor.go
    ├── plugin
    │   ├── init.go
    │   └── specialop.go
    ├── README.org
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
    │   └── state_test.go
    └── types
        ├── application.go
        ├── block.go
        ├── block_meta.go
        ├── canonical_json.go
        ├── common.go
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

This is directory structure of Angine, you can see that we have packed every module under its own directory. This give you a clear view of the parts making up an Angine.

1.  state/ is the running state of the Angine, which is driven by blockchain/, mempool/ and consensus/

2.  blockchain/ takes charge of syncing blocks, loading blocks, persisting blocks and anything that physically related to "block"

3.  mempool/ takes charge of buffering effective transactions and reaching an agreement about the order of transactions

4.  consensus/ takes charge of gossipping between peers, consensus algorithm related data stream


<a id="org3958c62"></a>

# What we have fulfilled

1.  CA based on asymmetric cyrpto

2.  Dynamically changing ValidatorSet of ConsensusState

3.  Two kinds of transactions, normal and special, are totally isolated. Special tx will only be processed by plugins.

4.  Angine plugins


<a id="org5e86273"></a>

# How to use


<a id="org9a260de"></a>

## Initialize Angine

This is how you initialize an angine. 

    angine.Initialize(&angine.AngineTunes{Conf: conf})

The "angine.Initialize" will handle the generation of default configs, genesis file and private key. You must only do this once for a particular chain, otherwise, your id might be different.


<a id="org19c7657"></a>

### Construct an AngineTunes.

    type AngineTunes struct {
        Runtime string
        Conf    *cfg.MapConfig
    }

This struct contains 2 fields and you only have to fill one:

1.  Runtime is a path that contains all the auto-generated files. So provided this will just generate everything under this path with random chainID and private/pub key pair.

2.  Conf contains anyconfig that you want to override the defaults. Say, you wanna use some cli args to override the default configs, this is the thing you should look into.

    After, you probably need to edit those files mannually to get exactly what you want. Everything in the files are straigt forward.


<a id="orgf101d9a"></a>

## New an Agnine instance and start it

First, you need to import angine into your project :-) then, 

    import "gitlab.zhonganonline.com/ann/angine"
    
    ...
    
    mainAngine := angine.NewAngine(&angine.AngineTunes{Conf: conf})
    
    ...
    
    mainAngine.Start()

That is all.

