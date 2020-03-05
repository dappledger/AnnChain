# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [chain/proto/grpc.proto](#chain/proto/grpc.proto)
    - [RpcService](#grpc.RpcService)
  
    - [CmdBlock](#grpc.CmdBlock)
    - [CmdBlockchainInfo](#grpc.CmdBlockchainInfo)
    - [CmdBroadcastTx](#grpc.CmdBroadcastTx)
    - [CmdChainID](#grpc.CmdChainID)
    - [CmdEventCode](#grpc.CmdEventCode)
    - [CmdHash](#grpc.CmdHash)
    - [CmdQuery](#grpc.CmdQuery)
    - [CmdRequestSpecialOP](#grpc.CmdRequestSpecialOP)
    - [EmptyRequest](#grpc.EmptyRequest)
  
  
  

- [gemmill/protos/blockchain/message.proto](#gemmill/protos/blockchain/message.proto)
  
    - [BlockHeaderRequestMessage](#blockchain.BlockHeaderRequestMessage)
    - [BlockHeaderResponseMessage](#blockchain.BlockHeaderResponseMessage)
    - [BlockMessage](#blockchain.BlockMessage)
    - [BlockRequestMessage](#blockchain.BlockRequestMessage)
    - [BlockResponseMessage](#blockchain.BlockResponseMessage)
    - [StatusRequestMessage](#blockchain.StatusRequestMessage)
    - [StatusResponseMessage](#blockchain.StatusResponseMessage)
  
    - [MsgType](#blockchain.MsgType)
  
  

- [gemmill/protos/consensus/message.proto](#gemmill/protos/consensus/message.proto)
  
    - [BlockPartMessage](#consensus.BlockPartMessage)
    - [CommitStepMessage](#consensus.CommitStepMessage)
    - [ConsensusMessage](#consensus.ConsensusMessage)
    - [HasVoteMessage](#consensus.HasVoteMessage)
    - [NewRoundStepMessage](#consensus.NewRoundStepMessage)
    - [ProposalMessage](#consensus.ProposalMessage)
    - [ProposalPOLMessage](#consensus.ProposalPOLMessage)
    - [ProtoBitArray](#consensus.ProtoBitArray)
    - [VoteMessage](#consensus.VoteMessage)
    - [VoteSetBitsMessage](#consensus.VoteSetBitsMessage)
    - [VoteSetMaj23Message](#consensus.VoteSetMaj23Message)
  
    - [MsgType](#consensus.MsgType)
    - [RoundStepType](#consensus.RoundStepType)
  
  

- [gemmill/protos/crypto/crypto.proto](#gemmill/protos/crypto/crypto.proto)
  
    - [PubKey](#crypto.PubKey)
    - [Signature](#crypto.Signature)
  
    - [KeyType](#crypto.KeyType)
  
  

- [gemmill/protos/events/events.proto](#gemmill/protos/events/events.proto)
  
    - [EventDataNewBlock](#events.EventDataNewBlock)
    - [EventDataNewBlockHeader](#events.EventDataNewBlockHeader)
    - [EventDataRoundState](#events.EventDataRoundState)
    - [EventDataTx](#events.EventDataTx)
    - [EventDataVote](#events.EventDataVote)
  
    - [EventType](#events.EventType)
  
  

- [gemmill/protos/mempool/message.proto](#gemmill/protos/mempool/message.proto)
  
    - [MempoolMessage](#mempool.MempoolMessage)
    - [TxMessage](#mempool.TxMessage)
  
    - [MsgType](#mempool.MsgType)
  
  

- [gemmill/protos/p2p/p2p.proto](#gemmill/protos/p2p/p2p.proto)
  
    - [ChannelStatus](#p2p.ChannelStatus)
    - [ConnectionStatus](#p2p.ConnectionStatus)
    - [FlowStatus](#p2p.FlowStatus)
    - [NodeInfo](#p2p.NodeInfo)
  
  
  

- [gemmill/protos/rpc/rpc.proto](#gemmill/protos/rpc/rpc.proto)
  
    - [Peer](#rpc.Peer)
    - [ResultBlock](#rpc.ResultBlock)
    - [ResultBlockchainInfo](#rpc.ResultBlockchainInfo)
    - [ResultBroadcastTx](#rpc.ResultBroadcastTx)
    - [ResultBroadcastTxCommit](#rpc.ResultBroadcastTxCommit)
    - [ResultCoreVersion](#rpc.ResultCoreVersion)
    - [ResultDialSeeds](#rpc.ResultDialSeeds)
    - [ResultDumpConsensusState](#rpc.ResultDumpConsensusState)
    - [ResultEvent](#rpc.ResultEvent)
    - [ResultGenesis](#rpc.ResultGenesis)
    - [ResultHealthInfo](#rpc.ResultHealthInfo)
    - [ResultInfo](#rpc.ResultInfo)
    - [ResultLastHeight](#rpc.ResultLastHeight)
    - [ResultNetInfo](#rpc.ResultNetInfo)
    - [ResultNonEmptyHeights](#rpc.ResultNonEmptyHeights)
    - [ResultNumArchivedBlocks](#rpc.ResultNumArchivedBlocks)
    - [ResultNumLimitTx](#rpc.ResultNumLimitTx)
    - [ResultOrgs](#rpc.ResultOrgs)
    - [ResultQuery](#rpc.ResultQuery)
    - [ResultRefuseList](#rpc.ResultRefuseList)
    - [ResultRequestAdminOP](#rpc.ResultRequestAdminOP)
    - [ResultShards](#rpc.ResultShards)
    - [ResultStatus](#rpc.ResultStatus)
    - [ResultSubscribe](#rpc.ResultSubscribe)
    - [ResultSurveillance](#rpc.ResultSurveillance)
    - [ResultTransaction](#rpc.ResultTransaction)
    - [ResultUnconfirmedTxs](#rpc.ResultUnconfirmedTxs)
    - [ResultUnsafeFlushMempool](#rpc.ResultUnsafeFlushMempool)
    - [ResultUnsafeProfile](#rpc.ResultUnsafeProfile)
    - [ResultUnsafeSetConfig](#rpc.ResultUnsafeSetConfig)
    - [ResultUnsubscribe](#rpc.ResultUnsubscribe)
    - [ResultValidator](#rpc.ResultValidator)
    - [ResultValidators](#rpc.ResultValidators)
  
    - [Type](#rpc.Type)
  
  

- [gemmill/protos/state/state.proto](#gemmill/protos/state/state.proto)
  
    - [GenesisDoc](#state.GenesisDoc)
    - [Plugin](#state.Plugin)
    - [QueryCachePlugin](#state.QueryCachePlugin)
    - [SpecialOp](#state.SpecialOp)
    - [State](#state.State)
    - [SuspectPlugin](#state.SuspectPlugin)
    - [ValidatorSet](#state.ValidatorSet)
  
    - [Type](#state.Type)
  
  

- [gemmill/protos/trace/message.proto](#gemmill/protos/trace/message.proto)
  
    - [TraceMessage](#trace.TraceMessage)
    - [TraceRequest](#trace.TraceRequest)
    - [TraceResponse](#trace.TraceResponse)
  
    - [MsgType](#trace.MsgType)
  
  

- [gemmill/protos/types/types.proto](#gemmill/protos/types/types.proto)
  
    - [Block](#types.Block)
    - [BlockID](#types.BlockID)
    - [BlockMeta](#types.BlockMeta)
    - [Commit](#types.Commit)
    - [Data](#types.Data)
    - [GenesisDoc](#types.GenesisDoc)
    - [GenesisValidator](#types.GenesisValidator)
    - [Header](#types.Header)
    - [Part](#types.Part)
    - [PartSetHeader](#types.PartSetHeader)
    - [Proposal](#types.Proposal)
    - [ProposalData](#types.ProposalData)
    - [Result](#types.Result)
    - [SignableBase](#types.SignableBase)
    - [SimpleProof](#types.SimpleProof)
    - [Transaction](#types.Transaction)
    - [TxData](#types.TxData)
    - [Validator](#types.Validator)
    - [ValidatorSet](#types.ValidatorSet)
    - [Vote](#types.Vote)
    - [VoteData](#types.VoteData)
  
    - [CodeType](#types.CodeType)
    - [VoteType](#types.VoteType)
  
  

- [Scalar Value Types](#scalar-value-types)



<a name="chain/proto/grpc.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## chain/proto/grpc.proto



<a name="grpc.RpcService"></a>

### RpcService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Status | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultStatus](#rpc.ResultStatus) |  |
| Genesis | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultGenesis](#rpc.ResultGenesis) |  |
| Health | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultHealthInfo](#rpc.ResultHealthInfo) |  |
| Block | [CmdBlock](#grpc.CmdBlock) | [.rpc.ResultBlock](#rpc.ResultBlock) |  |
| BlockchainInfo | [CmdBlockchainInfo](#grpc.CmdBlockchainInfo) | [.rpc.ResultBlockchainInfo](#rpc.ResultBlockchainInfo) |  |
| DumpConsensusState | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultDumpConsensusState](#rpc.ResultDumpConsensusState) |  |
| UnconfirmedTxs | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultUnconfirmedTxs](#rpc.ResultUnconfirmedTxs) |  |
| NumUnconfirmedTxs | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultUnconfirmedTxs](#rpc.ResultUnconfirmedTxs) |  |
| NumArchivedBlocks | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultNumArchivedBlocks](#rpc.ResultNumArchivedBlocks) |  |
| UnsafeFlushMempool | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultUnsafeFlushMempool](#rpc.ResultUnsafeFlushMempool) |  |
| BroadcastTx | [CmdBroadcastTx](#grpc.CmdBroadcastTx) | [.rpc.ResultBroadcastTx](#rpc.ResultBroadcastTx) |  |
| BroadcastTxCommit | [CmdBroadcastTx](#grpc.CmdBroadcastTx) | [.rpc.ResultBroadcastTxCommit](#rpc.ResultBroadcastTxCommit) |  |
| QueryTx | [CmdQuery](#grpc.CmdQuery) | [.rpc.ResultNumLimitTx](#rpc.ResultNumLimitTx) |  |
| Query | [CmdQuery](#grpc.CmdQuery) | [.rpc.ResultQuery](#rpc.ResultQuery) |  |
| GetTransactionByHash | [CmdHash](#grpc.CmdHash) | [.rpc.ResultQuery](#rpc.ResultQuery) |  |
| Info | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultInfo](#rpc.ResultInfo) |  |
| Validators | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultValidators](#rpc.ResultValidators) |  |
| CoreVersion | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultCoreVersion](#rpc.ResultCoreVersion) |  |
| LastHeight | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultLastHeight](#rpc.ResultLastHeight) |  |
| ZaSurveillance | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultSurveillance](#rpc.ResultSurveillance) |  |
| NetInfo | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultNetInfo](#rpc.ResultNetInfo) |  |
| Blacklist | [EmptyRequest](#grpc.EmptyRequest) | [.rpc.ResultRefuseList](#rpc.ResultRefuseList) |  |

 <!-- end services -->



<a name="grpc.CmdBlock"></a>

### CmdBlock



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |






<a name="grpc.CmdBlockchainInfo"></a>

### CmdBlockchainInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| MinHeight | [int64](#int64) |  |  |
| MaxHeight | [int64](#int64) |  |  |






<a name="grpc.CmdBroadcastTx"></a>

### CmdBroadcastTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Tx | [bytes](#bytes) |  |  |






<a name="grpc.CmdChainID"></a>

### CmdChainID



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ChainID | [string](#string) |  |  |






<a name="grpc.CmdEventCode"></a>

### CmdEventCode



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| CodeHash | [bytes](#bytes) |  |  |






<a name="grpc.CmdHash"></a>

### CmdHash



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Hash | [bytes](#bytes) |  |  |






<a name="grpc.CmdQuery"></a>

### CmdQuery



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Query | [bytes](#bytes) |  |  |






<a name="grpc.CmdRequestSpecialOP"></a>

### CmdRequestSpecialOP



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ChainID | [string](#string) |  |  |
| Tx | [bytes](#bytes) |  |  |






<a name="grpc.EmptyRequest"></a>

### EmptyRequest






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/blockchain/message.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/blockchain/message.proto


 <!-- end services -->



<a name="blockchain.BlockHeaderRequestMessage"></a>

### BlockHeaderRequestMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |






<a name="blockchain.BlockHeaderResponseMessage"></a>

### BlockHeaderResponseMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Header | [types.Header](#types.Header) |  |  |






<a name="blockchain.BlockMessage"></a>

### BlockMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Type | [MsgType](#blockchain.MsgType) |  |  |
| Data | [bytes](#bytes) |  |  |






<a name="blockchain.BlockRequestMessage"></a>

### BlockRequestMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |






<a name="blockchain.BlockResponseMessage"></a>

### BlockResponseMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Block | [types.Block](#types.Block) |  |  |






<a name="blockchain.StatusRequestMessage"></a>

### StatusRequestMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |






<a name="blockchain.StatusResponseMessage"></a>

### StatusResponseMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |





 <!-- end messages -->


<a name="blockchain.MsgType"></a>

### MsgType


| Name | Number | Description |
| ---- | ------ | ----------- |
| None | 0 |  |
| BlockReq | 1 |  |
| BlockRsp | 2 |  |
| StatusReq | 3 |  |
| StatusRsp | 4 |  |
| HeaderReq | 5 |  |
| HeaderRsp | 6 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/consensus/message.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/consensus/message.proto


 <!-- end services -->



<a name="consensus.BlockPartMessage"></a>

### BlockPartMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Part | [types.Part](#types.Part) |  |  |






<a name="consensus.CommitStepMessage"></a>

### CommitStepMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| BlockPartsHeader | [types.PartSetHeader](#types.PartSetHeader) |  |  |
| BlockParts | [ProtoBitArray](#consensus.ProtoBitArray) |  |  |






<a name="consensus.ConsensusMessage"></a>

### ConsensusMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Type | [MsgType](#consensus.MsgType) |  |  |
| Data | [bytes](#bytes) |  |  |






<a name="consensus.HasVoteMessage"></a>

### HasVoteMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Type | [types.VoteType](#types.VoteType) |  |  |
| Index | [int64](#int64) |  |  |






<a name="consensus.NewRoundStepMessage"></a>

### NewRoundStepMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Step | [RoundStepType](#consensus.RoundStepType) |  |  |
| SecondsSinceStartTime | [int64](#int64) |  |  |
| LastCommitRound | [int64](#int64) |  |  |






<a name="consensus.ProposalMessage"></a>

### ProposalMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Proposal | [types.Proposal](#types.Proposal) |  |  |






<a name="consensus.ProposalPOLMessage"></a>

### ProposalPOLMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| ProposalPOLRound | [int64](#int64) |  |  |
| ProposalPOL | [ProtoBitArray](#consensus.ProtoBitArray) |  |  |






<a name="consensus.ProtoBitArray"></a>

### ProtoBitArray



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Bits | [int64](#int64) |  |  |
| Elems | [uint64](#uint64) | repeated |  |






<a name="consensus.VoteMessage"></a>

### VoteMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Vote | [types.Vote](#types.Vote) |  |  |






<a name="consensus.VoteSetBitsMessage"></a>

### VoteSetBitsMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Type | [types.VoteType](#types.VoteType) |  |  |
| BlockID | [types.BlockID](#types.BlockID) |  |  |
| Votes | [ProtoBitArray](#consensus.ProtoBitArray) |  |  |






<a name="consensus.VoteSetMaj23Message"></a>

### VoteSetMaj23Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Type | [types.VoteType](#types.VoteType) |  |  |
| BlockID | [types.BlockID](#types.BlockID) |  |  |





 <!-- end messages -->


<a name="consensus.MsgType"></a>

### MsgType


| Name | Number | Description |
| ---- | ------ | ----------- |
| None | 0 |  |
| NewRoundStep | 1 |  |
| CommitStep | 2 |  |
| Proposal | 3 |  |
| ProposalPOL | 4 |  |
| BlockPart | 5 | both block & POL |
| Vote | 6 |  |
| HasVote | 7 |  |
| VoteSetMaj23 | 8 |  |
| VoteSetBits | 9 |  |



<a name="consensus.RoundStepType"></a>

### RoundStepType


| Name | Number | Description |
| ---- | ------ | ----------- |
| EnumBegin | 0 |  |
| NewHeight | 1 | Wait til CommitTime + timeoutCommit |
| NewRound | 2 | Setup new round and go to RoundStepPropose |
| Propose | 3 | Did propose, gossip proposal |
| Prevote | 4 | Did prevote, gossip prevotes |
| PrevoteWait | 5 | Did receive any +2/3 prevotes, start timeout |
| Precommit | 6 | Did precommit, gossip precommits |
| PrecommitWait | 7 | Did receive any +2/3 precommits, start timeout |
| Commit | 8 | Entered commit state machine |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/crypto/crypto.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/crypto/crypto.proto


 <!-- end services -->



<a name="crypto.PubKey"></a>

### PubKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bytes | [bytes](#bytes) |  |  |
| type | [KeyType](#crypto.KeyType) |  |  |






<a name="crypto.Signature"></a>

### Signature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bytes | [bytes](#bytes) |  |  |
| type | [KeyType](#crypto.KeyType) |  |  |





 <!-- end messages -->


<a name="crypto.KeyType"></a>

### KeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| Ed25519 | 0 |  |
| Secp256k1 | 1 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/events/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/events/events.proto


 <!-- end services -->



<a name="events.EventDataNewBlock"></a>

### EventDataNewBlock



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Block | [types.Block](#types.Block) |  |  |






<a name="events.EventDataNewBlockHeader"></a>

### EventDataNewBlockHeader



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Header | [types.Header](#types.Header) |  |  |






<a name="events.EventDataRoundState"></a>

### EventDataRoundState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Step | [string](#string) |  | RoundState is private |






<a name="events.EventDataTx"></a>

### EventDataTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Tx | [bytes](#bytes) |  |  |
| Data | [bytes](#bytes) |  |  |
| Log | [string](#string) |  |  |
| Code | [types.CodeType](#types.CodeType) |  |  |
| Error | [string](#string) |  |  |






<a name="events.EventDataVote"></a>

### EventDataVote



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Vote | [types.Vote](#types.Vote) |  |  |





 <!-- end messages -->


<a name="events.EventType"></a>

### EventType


| Name | Number | Description |
| ---- | ------ | ----------- |
| EventNewBlock | 0 |  |
| EventNewBlockHeader | 1 |  |
| EventTx | 2 |  |
| EventRoundState | 3 |  |
| EventVote | 4 |  |
| EventSwitchToConsensus | 5 |  |
| EventHookNewRound | 6 |  |
| EventHookPropose | 7 |  |
| EventHookPrecommit | 8 |  |
| EventHookCommit | 9 |  |
| EventHookExecute | 10 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/mempool/message.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/mempool/message.proto


 <!-- end services -->



<a name="mempool.MempoolMessage"></a>

### MempoolMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Type | [MsgType](#mempool.MsgType) |  |  |
| Data | [bytes](#bytes) |  |  |






<a name="mempool.TxMessage"></a>

### TxMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Tx | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="mempool.MsgType"></a>

### MsgType


| Name | Number | Description |
| ---- | ------ | ----------- |
| None | 0 |  |
| Tx | 1 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/p2p/p2p.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/p2p/p2p.proto


 <!-- end services -->



<a name="p2p.ChannelStatus"></a>

### ChannelStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ID | [uint32](#uint32) |  |  |
| SendQueueCapacity | [int32](#int32) |  |  |
| SendQueueSize | [int32](#int32) |  |  |
| Priority | [int32](#int32) |  |  |
| RecentlySent | [int64](#int64) |  |  |






<a name="p2p.ConnectionStatus"></a>

### ConnectionStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| SendMonitor | [FlowStatus](#p2p.FlowStatus) |  |  |
| RecvMonitor | [FlowStatus](#p2p.FlowStatus) |  |  |
| Channels | [ChannelStatus](#p2p.ChannelStatus) | repeated |  |






<a name="p2p.FlowStatus"></a>

### FlowStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Active | [bool](#bool) |  |  |
| Start | [int64](#int64) |  |  |
| Duration | [int64](#int64) |  |  |
| Idle | [int64](#int64) |  |  |
| Bytes | [int64](#int64) |  |  |
| Samples | [int64](#int64) |  |  |
| InstRate | [int64](#int64) |  |  |
| CurRate | [int64](#int64) |  |  |
| AvgRate | [int64](#int64) |  |  |
| PeakRate | [int64](#int64) |  |  |
| BytesRem | [int64](#int64) |  |  |
| TimeRem | [int64](#int64) |  |  |
| Progress | [uint32](#uint32) |  |  |






<a name="p2p.NodeInfo"></a>

### NodeInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| PubKey | [crypto.PubKey](#crypto.PubKey) |  |  |
| SigndPubKey | [string](#string) |  |  |
| Moniker | [string](#string) |  |  |
| Network | [string](#string) |  |  |
| RemoteAddr | [string](#string) |  |  |
| ListenAddr | [string](#string) |  |  |
| Version | [string](#string) |  |  |
| Other | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/rpc/rpc.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/rpc/rpc.proto


 <!-- end services -->



<a name="rpc.Peer"></a>

### Peer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| NodeInfo | [p2p.NodeInfo](#p2p.NodeInfo) |  |  |
| IsOutbound | [bool](#bool) |  |  |
| ConnectionStatus | [p2p.ConnectionStatus](#p2p.ConnectionStatus) |  |  |






<a name="rpc.ResultBlock"></a>

### ResultBlock



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| BlockMeta | [types.BlockMeta](#types.BlockMeta) |  |  |
| Block | [types.Block](#types.Block) |  |  |






<a name="rpc.ResultBlockchainInfo"></a>

### ResultBlockchainInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| LastHeight | [int64](#int64) |  |  |
| BlockMetas | [types.BlockMeta](#types.BlockMeta) | repeated |  |






<a name="rpc.ResultBroadcastTx"></a>

### ResultBroadcastTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Code | [types.CodeType](#types.CodeType) |  |  |
| Data | [bytes](#bytes) |  |  |
| TxHash | [string](#string) |  |  |
| Log | [string](#string) |  |  |






<a name="rpc.ResultBroadcastTxCommit"></a>

### ResultBroadcastTxCommit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Code | [types.CodeType](#types.CodeType) |  |  |
| Data | [bytes](#bytes) |  |  |
| TxHash | [string](#string) |  |  |
| Log | [string](#string) |  |  |






<a name="rpc.ResultCoreVersion"></a>

### ResultCoreVersion



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Version | [string](#string) |  |  |
| AppName | [string](#string) |  |  |
| AppVersion | [string](#string) |  |  |
| Hash | [string](#string) |  |  |






<a name="rpc.ResultDialSeeds"></a>

### ResultDialSeeds







<a name="rpc.ResultDumpConsensusState"></a>

### ResultDumpConsensusState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| RoundState | [string](#string) |  |  |
| PeerRoundStates | [string](#string) | repeated |  |






<a name="rpc.ResultEvent"></a>

### ResultEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Name | [string](#string) |  |  |
| Type | [events.EventType](#events.EventType) |  |  |
| Event | [bytes](#bytes) |  |  |






<a name="rpc.ResultGenesis"></a>

### ResultGenesis



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Genesis | [types.GenesisDoc](#types.GenesisDoc) |  |  |






<a name="rpc.ResultHealthInfo"></a>

### ResultHealthInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Status | [int32](#int32) |  |  |






<a name="rpc.ResultInfo"></a>

### ResultInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Data | [string](#string) |  |  |
| Version | [string](#string) |  |  |
| LastBlockHeight | [int64](#int64) |  |  |
| LastBlockAppHash | [bytes](#bytes) |  |  |






<a name="rpc.ResultLastHeight"></a>

### ResultLastHeight



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| LastHeight | [int64](#int64) |  |  |






<a name="rpc.ResultNetInfo"></a>

### ResultNetInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Listening | [bool](#bool) |  |  |
| Listeners | [string](#string) | repeated |  |
| peers | [Peer](#rpc.Peer) | repeated |  |






<a name="rpc.ResultNonEmptyHeights"></a>

### ResultNonEmptyHeights



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Heights | [int64](#int64) | repeated |  |






<a name="rpc.ResultNumArchivedBlocks"></a>

### ResultNumArchivedBlocks



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Num | [int64](#int64) |  |  |






<a name="rpc.ResultNumLimitTx"></a>

### ResultNumLimitTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Num | [uint64](#uint64) |  |  |






<a name="rpc.ResultOrgs"></a>

### ResultOrgs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Names | [string](#string) | repeated |  |






<a name="rpc.ResultQuery"></a>

### ResultQuery



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Result | [types.Result](#types.Result) |  |  |






<a name="rpc.ResultRefuseList"></a>

### ResultRefuseList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Result | [string](#string) | repeated |  |






<a name="rpc.ResultRequestAdminOP"></a>

### ResultRequestAdminOP



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Code | [types.CodeType](#types.CodeType) |  |  |
| Data | [bytes](#bytes) |  |  |
| Log | [string](#string) |  |  |






<a name="rpc.ResultShards"></a>

### ResultShards



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Names | [string](#string) | repeated |  |






<a name="rpc.ResultStatus"></a>

### ResultStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| NodeInfo | [p2p.NodeInfo](#p2p.NodeInfo) |  |  |
| PubKey | [crypto.PubKey](#crypto.PubKey) |  |  |
| LatestBlockHash | [bytes](#bytes) |  |  |
| LatestAppHash | [bytes](#bytes) |  |  |
| LatestBlockHeight | [int64](#int64) |  |  |
| LatestBlockTime | [int64](#int64) |  |  |






<a name="rpc.ResultSubscribe"></a>

### ResultSubscribe







<a name="rpc.ResultSurveillance"></a>

### ResultSurveillance



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| NanoSecsPerTx | [int64](#int64) |  |  |
| Height | [int64](#int64) |  |  |
| Addr | [string](#string) |  |  |
| IsValidator | [bool](#bool) |  |  |
| NumValidators | [int64](#int64) |  |  |
| NumPeers | [int64](#int64) |  |  |
| RunningTime | [int64](#int64) |  |  |
| PubKey | [string](#string) |  |  |






<a name="rpc.ResultTransaction"></a>

### ResultTransaction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| BlockHash | [bytes](#bytes) |  |  |
| BlockHeight | [uint64](#uint64) |  |  |
| TransactionIndex | [uint64](#uint64) |  |  |
| RawTransaction | [bytes](#bytes) |  |  |
| Timestamp | [uint64](#uint64) |  |  |






<a name="rpc.ResultUnconfirmedTxs"></a>

### ResultUnconfirmedTxs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| N | [int64](#int64) |  |  |
| Txs | [bytes](#bytes) | repeated |  |






<a name="rpc.ResultUnsafeFlushMempool"></a>

### ResultUnsafeFlushMempool







<a name="rpc.ResultUnsafeProfile"></a>

### ResultUnsafeProfile







<a name="rpc.ResultUnsafeSetConfig"></a>

### ResultUnsafeSetConfig







<a name="rpc.ResultUnsubscribe"></a>

### ResultUnsubscribe







<a name="rpc.ResultValidator"></a>

### ResultValidator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Address | [bytes](#bytes) |  |  |
| PubKey | [string](#string) |  |  |
| VotingPower | [int64](#int64) |  |  |
| Accum | [int64](#int64) |  |  |
| IsCA | [bool](#bool) |  |  |






<a name="rpc.ResultValidators"></a>

### ResultValidators



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| BlockHeight | [int64](#int64) |  |  |
| Validators | [ResultValidator](#rpc.ResultValidator) | repeated |  |





 <!-- end messages -->


<a name="rpc.Type"></a>

### Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| RpcNone | 0 |  |
| RpcGenesis | 1 |  |
| RpcBlockchainInfo | 2 |  |
| RpcBlock | 3 |  |
| RpcNonEmptyHeights | 4 |  |
| RpcStatus | 32 | 0x2 bytes are for the network |
| RpcNetInfo | 33 |  |
| RpcDialSeeds | 34 |  |
| RpcOrgs | 35 |  |
| RpcRefuseList | 16 | 0x1 bytes are for refuseList |
| RpcValidators | 64 | 0x4 bytes are for the consensus |
| RpcDumpConsensusState | 65 |  |
| RpcBroadcastTx | 96 | 0x6 bytes are for txs / the application |
| RpcUnconfirmedTxs | 97 |  |
| RpcBroadcastTxCommit | 98 |  |
| RpcRequestSpecialOP | 99 |  |
| RpcQuery | 112 | 0x7 bytes are for querying the application |
| RpcInfo | 113 |  |
| RpcSubscribe | 128 | 0x8 bytes are for events |
| RpcUnsubscribe | 129 |  |
| RpcEvent | 130 |  |
| RpcUnsafeSetConfig | 160 | 0xa bytes for testing |
| RpcUnsafeStartCPUProfiler | 161 |  |
| RpcUnsafeStopCPUProfiler | 162 |  |
| RpcUnsafeWriteHeapProfile | 163 |  |
| RpcUnsafeFlushMempool | 164 |  |
| RpcCoreVersion | 175 |  |
| RpcSurveillance | 144 | 0x9 bytes are for za_surveillance |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/state/state.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/state/state.proto


 <!-- end services -->



<a name="state.GenesisDoc"></a>

### GenesisDoc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| JSONData | [bytes](#bytes) |  |  |






<a name="state.Plugin"></a>

### Plugin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Type | [Type](#state.Type) |  |  |
| PData | [bytes](#bytes) |  |  |






<a name="state.QueryCachePlugin"></a>

### QueryCachePlugin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| JSONData | [bytes](#bytes) |  |  |






<a name="state.SpecialOp"></a>

### SpecialOp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| JSONData | [bytes](#bytes) |  |  |






<a name="state.State"></a>

### State



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| GenesisDoc | [GenesisDoc](#state.GenesisDoc) |  |  |
| ChainID | [string](#string) |  |  |
| LastBlockHeight | [int64](#int64) |  |  |
| LastBlockID | [types.BlockID](#types.BlockID) |  |  |
| LastBlockTime | [int64](#int64) |  |  |
| Validators | [ValidatorSet](#state.ValidatorSet) |  |  |
| LastValidators | [ValidatorSet](#state.ValidatorSet) |  |  |
| LastNonEmptyHeight | [int64](#int64) |  |  |
| AppHash | [bytes](#bytes) |  |  |
| ReceiptsHash | [bytes](#bytes) |  | repeated Plugin Plugins = 11; |






<a name="state.SuspectPlugin"></a>

### SuspectPlugin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| JSONData | [bytes](#bytes) |  |  |






<a name="state.ValidatorSet"></a>

### ValidatorSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| JSONData | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="state.Type"></a>

### Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| PluginNone | 0 |  |
| PluginSpecialOp | 1 |  |
| PluginSuspect | 2 |  |
| PluginQueryCache | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/trace/message.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/trace/message.proto


 <!-- end services -->



<a name="trace.TraceMessage"></a>

### TraceMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Type | [MsgType](#trace.MsgType) |  |  |
| Data | [bytes](#bytes) |  |  |






<a name="trace.TraceRequest"></a>

### TraceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Data | [bytes](#bytes) |  |  |






<a name="trace.TraceResponse"></a>

### TraceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| RequestHash | [bytes](#bytes) |  |  |
| Resp | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="trace.MsgType"></a>

### MsgType


| Name | Number | Description |
| ---- | ------ | ----------- |
| None | 0 |  |
| Request | 1 |  |
| Responce | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="gemmill/protos/types/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gemmill/protos/types/types.proto


 <!-- end services -->



<a name="types.Block"></a>

### Block



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Header | [Header](#types.Header) |  |  |
| Data | [Data](#types.Data) |  |  |
| LastCommit | [Commit](#types.Commit) |  |  |






<a name="types.BlockID"></a>

### BlockID



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Hash | [bytes](#bytes) |  |  |
| PartsHeader | [PartSetHeader](#types.PartSetHeader) |  |  |






<a name="types.BlockMeta"></a>

### BlockMeta



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Hash | [bytes](#bytes) |  |  |
| Header | [Header](#types.Header) |  |  |
| PartsHeader | [PartSetHeader](#types.PartSetHeader) |  |  |






<a name="types.Commit"></a>

### Commit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| BlockID | [BlockID](#types.BlockID) |  |  |
| Precommits | [Vote](#types.Vote) | repeated |  |






<a name="types.Data"></a>

### Data



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Txs | [bytes](#bytes) | repeated |  |
| ExTxs | [bytes](#bytes) | repeated |  |
| Hash | [bytes](#bytes) |  |  |






<a name="types.GenesisDoc"></a>

### GenesisDoc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| GenesisTime | [int64](#int64) |  |  |
| ChainID | [string](#string) |  |  |
| Validators | [GenesisValidator](#types.GenesisValidator) | repeated |  |
| AppHash | [bytes](#bytes) |  |  |
| Plugins | [string](#string) |  |  |






<a name="types.GenesisValidator"></a>

### GenesisValidator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Pubkey | [crypto.PubKey](#crypto.PubKey) |  |  |
| Amount | [int64](#int64) |  |  |
| Name | [string](#string) |  |  |
| IsCA | [bool](#bool) |  |  |
| RPCAddress | [string](#string) |  |  |






<a name="types.Header"></a>

### Header



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ChainID | [string](#string) |  |  |
| Height | [int64](#int64) |  |  |
| Time | [int64](#int64) |  |  |
| NumTxs | [int64](#int64) |  |  |
| Maker | [bytes](#bytes) |  |  |
| LastBlockID | [BlockID](#types.BlockID) |  |  |
| LastCommitHash | [bytes](#bytes) |  |  |
| DataHash | [bytes](#bytes) |  |  |
| ValidatorsHash | [bytes](#bytes) |  |  |
| AppHash | [bytes](#bytes) |  |  |
| ReceiptsHash | [bytes](#bytes) |  |  |
| ProposerAddress | [bytes](#bytes) |  |  |
| Extra | [bytes](#bytes) |  |  |






<a name="types.Part"></a>

### Part



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Index | [int32](#int32) |  |  |
| Bytes | [bytes](#bytes) |  |  |
| Proof | [SimpleProof](#types.SimpleProof) |  |  |






<a name="types.PartSetHeader"></a>

### PartSetHeader



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Total | [int32](#int32) |  |  |
| Hash | [bytes](#bytes) |  |  |






<a name="types.Proposal"></a>

### Proposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Data | [ProposalData](#types.ProposalData) |  |  |
| Signature | [bytes](#bytes) |  |  |






<a name="types.ProposalData"></a>

### ProposalData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| BlockPartsHeader | [PartSetHeader](#types.PartSetHeader) |  |  |
| POLRound | [int64](#int64) |  |  |
| POLBlockID | [BlockID](#types.BlockID) |  |  |






<a name="types.Result"></a>

### Result



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Code | [CodeType](#types.CodeType) |  |  |
| Data | [bytes](#bytes) |  |  |
| Log | [string](#string) |  |  |






<a name="types.SignableBase"></a>

### SignableBase



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ChainID | [string](#string) |  |  |
| Data | [bytes](#bytes) |  |  |






<a name="types.SimpleProof"></a>

### SimpleProof



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bytes | [bytes](#bytes) | repeated |  |






<a name="types.Transaction"></a>

### Transaction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Data | [TxData](#types.TxData) |  |  |
| Pubkey | [crypto.PubKey](#crypto.PubKey) |  |  |
| Signature | [bytes](#bytes) |  |  |






<a name="types.TxData"></a>

### TxData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Nonce | [uint64](#uint64) |  |  |
| Recipient | [bytes](#bytes) |  |  |
| Amount | [bytes](#bytes) |  |  |
| Payload | [bytes](#bytes) |  |  |






<a name="types.Validator"></a>

### Validator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Address | [bytes](#bytes) |  |  |
| Pubkey | [crypto.PubKey](#crypto.PubKey) |  |  |
| VotingPower | [int64](#int64) |  |  |
| Accum | [int64](#int64) |  |  |
| IsCA | [bool](#bool) |  |  |






<a name="types.ValidatorSet"></a>

### ValidatorSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| VSet | [Validator](#types.Validator) | repeated |  |






<a name="types.Vote"></a>

### Vote



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ValidatorAddress | [bytes](#bytes) |  |  |
| ValidatorIndex | [int32](#int32) |  |  |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Type | [VoteType](#types.VoteType) |  |  |
| BlockID | [BlockID](#types.BlockID) |  |  |
| Signature | [crypto.Signature](#crypto.Signature) |  |  |






<a name="types.VoteData"></a>

### VoteData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ValidatorAddress | [bytes](#bytes) |  |  |
| ValidatorIndex | [int64](#int64) |  |  |
| Height | [int64](#int64) |  |  |
| Round | [int64](#int64) |  |  |
| Type | [VoteType](#types.VoteType) |  |  |
| BlockID | [BlockID](#types.BlockID) |  |  |





 <!-- end messages -->


<a name="types.CodeType"></a>

### CodeType


| Name | Number | Description |
| ---- | ------ | ----------- |
| OK | 0 |  |
| InternalError | 1 | General response codes, 0 ~ 99 |
| EncodingError | 2 |  |
| BadNonce | 3 |  |
| Unauthorized | 4 |  |
| InsufficientFunds | 5 |  |
| UnknownRequest | 6 |  |
| InvalidTx | 7 |  |
| BaseDuplicateAddress | 101 | Reserved for basecoin, 100 ~ 199 |
| BaseEncodingError | 102 |  |
| BaseInsufficientFees | 103 |  |
| BaseInsufficientFunds | 104 |  |
| BaseInsufficientGasPrice | 105 |  |
| BaseInvalidInput | 106 |  |
| BaseInvalidOutput | 107 |  |
| BaseInvalidPubKey | 108 |  |
| BaseInvalidSequence | 109 |  |
| BaseInvalidSignature | 110 |  |
| BaseUnknownAddress | 111 |  |
| BaseUnknownPubKey | 112 |  |
| BaseUnknownPlugin | 113 |  |
| WrongRLP | 114 |  |
| SaveFailed | 115 |  |
| GovUnknownEntity | 201 | Reserved for governance, 200 ~ 299 |
| GovUnknownGroup | 202 |  |
| GovUnknownProposal | 203 |  |
| GovDuplicateGroup | 204 |  |
| GovDuplicateMember | 205 |  |
| GovDuplicateProposal | 206 |  |
| GovDuplicateVote | 207 |  |
| GovInvalidMember | 208 |  |
| GovInvalidVote | 209 |  |
| GovInvalidVotingPower | 210 |  |



<a name="types.VoteType"></a>

### VoteType


| Name | Number | Description |
| ---- | ------ | ----------- |
| None | 0 |  |
| Prevote | 1 |  |
| Precommit | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->


## Scalar Value Types

| .proto Type | Notes | Go Type | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |----------|
| <a name="double" /> double |  | float64 | double | double | float |
| <a name="float" /> float |  | float32 | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | []byte | string | ByteString | str |
