<div id="table-of-contents">
<h2>Table of Contents</h2>
<div id="text-table-of-contents">
<ul>
<li><a href="#orge8d1e56">1. What is angine?</a></li>
<li><a href="#orgb549c00">2. Structure of Engine</a></li>
<li><a href="#org1128696">3. What we have fulfilled</a></li>
</ul>
</div>
</div>

<a id="orge8d1e56"></a>

# What is angine?

Angine is a completely self-contained blockchain consensus engine. 
At the core, we use Tendermint BFT. It is an implementation of a Byzantine Fault Tolerant PoS consensus algorithm. So thank you Tendermint team.
We just wrap those many things that tendermint already offered together under a concept, "Engine". 


<a id="orgb549c00"></a>

# Structure of Engine

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
    ├── engine.go
    ├── engine_test.go
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

This is directory structure of Engine, you can see that we have packed every module under its own directory. This give you a clear view of the parts making up an Engine.

1.  state/ is the running state of the Engine, which is driven by blockchain/, mempool/ and consensus/

2.  blockchain/ takes charge of syncing blocks, loading blocks, persisting blocks and anything that physically related to "block"

3.  mempool/ takes charge of buffering effective transactions and reaching an agreement about the order of transactions

4.  consensus/ takes charge of gossipping between peers, consensus algorithm related data stream


<a id="org1128696"></a>

# What we have fulfilled

1.  CA based on asymmetric cyrpto

2.  Dynamically changing ValidatorSet of ConsensusState

