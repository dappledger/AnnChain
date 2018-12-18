## 错误码清单

### General response codes, 0 ~ 99

| Name                   | CodeType | Message    |
| ---------------------- | -------- | ---------- |
| CodeType_OK            | 0        | OK         |
| CodeType_InternalError | 1        | 内部错误   |
| CodeType_BadNonce      | 2        | 无效Nonce  |
| CodeType_InvalidTx     | 3        | 无效TX     |
| CodeType_LowBalance    | 4        | 余额不足   |
| CodeType_Timeout       | 5        | 交易超时   |
| CodeType_NullData      | 6        | 无数据返回 |
| CodeType_DecodingError | 7        | 解码错误   |
| CodeType_EncodingError | 8        | 编码错误   |

### Reserved for basecoin, 100 ~ 199

| Name                           | CodeType | Message             |
| ------------------------------ | -------- | ------------------- |
| CodeType_BaseInsufficientFunds | 101      | 手续费不足          |
| CodeType_BaseInvalidInput      | 102      | 无效的payload       |
| CodeType_BaseInvalidSignature  | 103      | 无效的签名          |
| CodeType_BaseUnknownAddress    | 104      | 未知地址            |
| CodeType_WrongRLP              | 105      | RLP编码错误         |
| CodeType_SaveFailed            | 106      | manage_data入库失败 |
|                                |          |                     |

### Reserved for contract, 400 ~ 499

| Name               | CodeType | message       |
| ------------------ | -------- | ------------- |
| CodeType_BadLimit  | 401      | gas超过上限   |
| CodeType_BadPrice  | 402      | 无效的Price   |
| CodeType_BadAmount | 403      | 无效的gas数量 |
|                    |          |               |

### Response Error for RPC 

| CodeType | message           |
| -------- | ----------------- |
| -32600   | JSON解码错误      |
| -32000   | 无效JSONRPC客户端 |
| -32601   | RPC方法未知       |
| -32001   | 方法只限于WS协议  |
| -32700   | JSON参数转换失败  |
| -32701   | http参数转换失败  |

