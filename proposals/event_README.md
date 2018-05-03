## App要求

* app需要embed EventAppBase。

  ```go
  type EVMApp struct {
  	.
    	.
    	.
  	civil.EventAppBase

  	core civil.Core
    	.
    	.
    	.
  }

  app := &EVMApp{
  		.
  		.
    		.
  		EventAppBase:      civil.NewEventAppBase(logger, config.GetString("cosi_laddr")),
    		.
    		.
    		.
  	}

  func (app *EVMApp) SetCore(core civil.Core) {
  	app.core = core
  	app.EventAppBase.SetCore(core)
  }

  func (app *EVMApp) Start() (err error) {
  	.
    	.
    	.
  	if _, err := app.EventAppBase.Start(); err != nil {
  		app.Stop()
  		return errors.Wrap(err, "[EVMAPP Start]")
  	}
  	.
    	.
    	.
  	return nil
  }

  func (app *EVMApp) Stop() {
  	.
    	.
    	.
  	app.EventAppBase.Stop()
  	.
    	.
    	.
  }
  ```

  ​

## 操作流程
1. 在被监听组织和监听组织上各自部署相关代码
2. 监听组织发起request请求，请求获得订阅事件的资格，在该请求中指定监听组织和被监听组织各自应用的代码。
3. 由request请求向被监听组织发起一个多签名的交易，由被监听组织对请求者的ID和指定的代码进行一轮投票。
4. 如果投票通过，被监听组织发起一个subscribe交易，广播确认建立请求的事件订阅关系。由全网节点共同记录事件订阅关系和使用的相关代码。如果投票失败，本次request请求的流程中断。不再进行后续操作。
5. 建立好订阅关系之后，每次被监听组织出块的时候，由被监听组织自行决定事件的数据，把数据在本地存储，然后如果目前存在N个事件订阅者，就向主链发送N个事件通知交易。附带事件的ID信息。
6. 订阅者解析交易，获得通知内的事件相关信息，然后由本组织内的某一个特定节点向被监听组织中的某一个特定节点发起tcp连接。
7. 通过连接，发送节点的身份认证信息，然后获取事件数据。
8. 事件数据根据双方的链下约定，经过各自的制定代码解析，完成事件数据的传输。
9. 双发在各自的组织内，进行事件发送完成的相关确认工作。

## 相关代码
1. lua代码

   事件发送和监听端所部属的lua代码，都必须以一个单独的main函数部署。函数只有一个入参和一个返回值，分别代表了外部传入的参数数据和处理后返回的有效数据。

   1. return nil：表示不需要继续执行event的相关流程，将不会再有event notification或者handle event被执行。相关的数据被直接丢弃。
   2. return string：表示脚本执行出现了错误，返回一个错误信息。
   3. return table：一个lua table包含了所有的event有效数据，key value形式，将被序列化。

   例如：

   ```lua
   function main(params)
   if params.contract_call == nil or params.contract_call["function"] ~= "buyChicken" then
     return nil;
   end

   ret = {};
   ret["from"] = params.from;
   ret["to"] = params.to;
   ret["value"] = params.value;
   ret["nonce"] = params.nonce;
   ret["function"] = params.contract_call["function"];
   ret["score"] = params.contract_call["_score"];

   return ret;
   end
   ```

   ​
