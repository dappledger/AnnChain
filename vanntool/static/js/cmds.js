function SpecailChangeValidator(){
	this.backend = "V节点RPC监听地址(<IP>:<Port>) 或者 <V节点名字>";
	this.privkey = "V节点私钥，如果backend填名字，这里填私钥密码";
	this.to_v_node = "要加入validator_set的节点的公钥 或者 <节点的名字>?<节点私钥密码>";
	this.target = "接受命令的组织名称(链名)";
	this.power = "validator节点的权重";
	this.isCA = true;
}

function Sign(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.sec = "加密使用的私钥，如果backend填名字，这里填私钥密码";
	this.pub = "明文";
}

function OrgCreate(){
	this.title = 'CREATE';
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称(链名)，通常是基础组织名称";
	this.orgid = "要创建的组织名称";
	this.app_list= "新组织加载的应用名称";
	this.p2p_laddr = "新组织的监听端口，格式：<Port>";
	this.seeds = "新组织可以连接的其他节点监听地址(<IP>:<Port>) 或者 <节点名字>:<Port>，英文逗号分割";
	this.sign_by_CA = "CA节点对<本节点公钥><新组织名称>的签名结果 或者 <CA节点的名字>?<CA节点私钥密码>";
	this.genesisnode = "初始成员节点，格式：<节点名字>?<节点密码>(amount:100,is_ca:true);<节点名字>?<节点密码>" ;
	this.configfile = "~/.ann_runtime/config.toml";
	this.genesisfile = "~/.ann_runtime/genesisfile.json";
}

function OrgJoin(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称，通常是基础组织名称";
	this.orgid = "要加入的组织名称";
	this.app_list= "新组织加载的应用名称";
	this.p2p_laddr = "新组织的监听端口，格式：50000";
	this.seeds = "新组织可以连接的其他节点监听地址(<IP>:<Port>) 或者 <节点名称>:<Port>，多个以英文逗号分割";
	this.sign_by_CA = "CA节点对<本节点公钥><新组织名称>的签名结果 或者 <CA节点的名字>?<CA节点私钥密码>";
	this.genesisnode = "初始成员节点，如果genesisnode中没有要加入节点的信息，这里可不填。格式：<节点名字>?<节点密码>(amount:100,is_ca:true);<节点名字>?<节点密码>" 
	this.configfile = "~/.ann_runtime/config.toml";
	//this.genesisfile = "~/.ann_runtime/genesisfile.json";
}

function OrgLeave(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.orgid = "要离开的组织名称";
}

function EventUploadCode(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称，通常是基础组织名称";
	this.code_text = "事件触发执行的lua code";
	this.ownerid = "代码所属的组织名称";
}

function EventRequest(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称，通常是基础组织名称";
	this.source = "事件触发组织的名称";
	this.source_hash = "事件要触发的code_hash";
	this.listener = "监听节点的组织名称";
	this.listener_hash = "监听节点收到事件触发的code_hash";
}

function EventUnsubscribe(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	this.privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称，通常是基础组织名称";
	this.listener = "监听节点的组织名称";
	this.source = "事件触发组织的名称";
}

function EvmCreate(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	//this.node_privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称(链名)";
	this.privkey = "账户私钥";
	this.params = "合约构造函数参数";
	this.abi_definition_text = "合约接口描述";
	this.code_text = "合约编译后的字节码";
}

function EvmCallOrRead(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	//this.node_privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称(链名)";
	this.privkey = "账户私钥";
	this.contract = "合约地址";
	this.method = "evm执行的函数名";
	this.params = "调用的函数参数";
	this.abi_definition_text = "合约接口描述";
}

function IkhofiCreate(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	//this.node_privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称(链名)";
	//this.contract_file = "合约路径";
	this.privkey = "账户私钥";
	this.contractid= "新合约名字";
}

function IkhofiCall(){
	this.backend = "节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>";
	//this.node_privkey = "节点私钥，如果backend填名字，这里填私钥密码";
	this.target = "接受命令的组织名称(链名)";
	this.privkey = "账户私钥";
	this.contractid = "合约名字";
	this.method = "ikhofi执行的函数名";
	this.params = "ikhofi函数执行参数('参数1','参数2',..)";
}

function CreateCmd(cmd , op){
	var cmdObj = new Object;
	switch (cmd){
		case "special":
			switch (op){
				case "change_validator"	:
					cmdObj = (new SpecailChangeValidator());
					break;
			}
			break;
		case "sign":
			cmdObj = (new Sign());
			break;
		case "organization":
			switch(op){
				case "create":
					cmdObj = (new OrgCreate());
					break;
				case "join":
					cmdObj = (new OrgJoin());
					break;
				case "leave":
					cmdObj = (new OrgLeave());
					break;
			}
			break;
		case "event":
			switch(op){
				case "uploadcode":
					cmdObj = (new EventUploadCode());
					break;
				case "request":
					cmdObj = (new EventRequest());
					break;
				case "unsubscribe":
					cmdObj = (new EventUnsubscribe());
					break;
			}
			break;
		case "evm":
			switch(op){
				case "create":
					cmdObj = (new EvmCreate());
					break;
				case "call":
				case "read":
					cmdObj = (new EvmCallOrRead());
					break;
			}
			break;
		case "jvm":
			switch(op){
				case "create":
					cmdObj = (new IkhofiCreate());
					break
				case "call":
				case "query":
					cmdObj = (new IkhofiCall());
					break;
			}
			break;
	}
	return cmdObj;
}

function parseTableStyle(label, inner){
	return   "<div class='row cl'><label class='form-label col-xs-2'>"+label+"</label><div class='formControls col-xs-8'>"+inner+"</div></div>";
}

function acceptFile(file){
	if (file.indexOf("config") > -1){
		return ".toml";
	}
//	if (file.indexOf("genesis") > -1){
//		return ".json";
//	}
	return ".json";
}

function innerHTML(type, key, defvalue){
	var str = "";
	switch(type){
		case "file":
			str = "<span class='btn-upload form-group'><input class='input-text upload-url' name='"+key+"_dir' id='"+key+"_dir' style='width:200px'><a href='javascript:void();' class='btn btn-primary upload-btn'><i class='Hui-iconfont'>&#xe642;</i> 浏览文件</a><input type='file' name='"+key+"_file' id='"+key+"_file' accept='"+acceptFile(key)+"' class='input-file' onchange='ChangeFile(\""+key+"\", this.value, this.files)'></span>";
			break;
		case "bool":
			str = "<div class='radio-box'><input type='radio' id='"+key+"_T' name='"+key+"' checked><label for='"+key+"_T'>true</label></div><div class='radio-box'><input type='radio' id='"+key+"_F' name='"+key+"'><label for='"+key+"_F'>false</label></div>";
			break;
		case "text":
			str = "<textarea name='"+key+"' id='"+key+"' cols=40 rows=4></textarea>";
			break;
		case "list":
			// TODO only support appnames_list now, want to fit more types
			str = "<input list='_appnames_list' name='"+key+"' id='"+key+"'> <datalist id='_appnames_list'><option label='evm' value='evm'><option label='ikhofi' value='ikhofi'><option label='noop' value='noop'><option label='remote' value='remote'></datalist>";
			break
		default:
			str = "<input type='text' class='input-text' placeholder='"+defvalue+"' name='"+key+"' id='"+key+"' autocomplete='off'>";
	}
	return str; 
}

function parseType(key){
	if (key.indexOf("file") > -1){
		return "file";
	}
	if (key.indexOf("_list") > -1){
		return "list";
	}
	if (key.indexOf("_text") > -1){
		return "text"
	}
	if (key.indexOf("is") == 0){
		return "bool";
	}
	return key;
}

function jointHTML(key, value){
	var keytype = parseType(key);
	switch (keytype){
		case "file":
			str = parseTableStyle(key+"：", innerHTML(keytype, key, value)) + parseTableStyle("", innerHTML("text", key, value));
			break;
		case "list":
		case "bool":
		case "text":
			//str = parseTableStyle(key+"：", innerHTML(keytype, key, value));
			//break;
		default:
			str = parseTableStyle(key+"：", innerHTML(keytype, key, value));
	}
	return str;
}

function AddTable(cmdObj){
	var html = "";
	// html=jointHTML("backend", "tcp://0.0.0.0:46657");
	for (var key in cmdObj)	{
		if (typeof(cmdObj[key]) == "function"){
			continue;	
		}
		html = html + jointHTML(key, cmdObj[key]);
		continue;
	}
	return html;
}

function ChangeFile(keyid, value, file){                                 
	document.getElementById(keyid+"_dir").value = value;                  
	ShowFile(keyid, file[0]);
}                                                                  

function ShowFile(keyid, file){
	var reader = new FileReader();
	reader.onload = function(progressEvent){
		document.getElementById(keyid).value = this.result;
	};
	reader.readAsText(file);
	return reader.result;
}

function GetRequest() {
	var url = location.search; //获取url中"?"符后的字串
	var theRequest = new Object();
	if (url.indexOf("?") != -1) {
		var str = url.substr(1);
		strs = str.split("&");
		for(var i = 0; i < strs.length; i ++) {
			theRequest[strs[i].split("=")[0]]=unescape(strs[i].split("=")[1]);
		}
	}
	return theRequest;
}
