function SpecailChangeValidator() {
	this.backend = "'V节点RPC监听地址(<IP>:<Port>) 或者 <V节点名字>'|input";
	this.privkey = "'V节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.to_v_node = "'要加入validator_set的节点的公钥 或者 <节点的名字>?<节点私钥密码>'|input";
	this.target = "'接受命令的组织名称(链名)'|input";
	this.power = "'validator节点的权重'|input";
	this.isCA = "|checkbox";
}

function Sign() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.sec = "'加密使用的私钥，如果backend填名字，这里填私钥密码'|input";
	this.pub = "待签名的节点公钥|input";
}

function OrgCreate(){
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.privkey = "'节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.target = "'接受命令的组织名称(链名)，通常是基础组织名称'|input";
	this.orgid = "'要创建的组织名称'|input";
	this.app_list= "'新组织加载的应用名称'|input";
	this.p2p_laddr = "'新组织的监听端口，格式：<Port>'|input";
	this.seeds = "'新组织可以连接的其他节点监听地址(<IP>:<Port>) 或者 <节点名字>:<Port>，英文逗号分割'|input";
	this.sign_by_CA = "'CA节点对<本节点公钥><新组织名称>的签名结果 或者 <CA节点的名字>?<CA节点私钥密码>'|input";
	this.genesisnode = "'初始成员节点，格式：<节点名字>?<节点密码>(amount:100,is_ca:true);<节点名字>?<节点密码>'|input";
	this.configfile = "'~/.ann_runtime/config.toml'|file";
	this.genesisfile = "'~/.ann_runtime/genesisfile.json'|file";
}

function OrgJoin() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.privkey = "'节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.target = "'接受命令的组织名称，通常是基础组织名称'|input";
	this.orgid = "要加入的组织名称|input";
	this.app_list = "新组织加载的应用名称|input";
	this.p2p_laddr = "'新组织的监听端口，格式：50000'|input";
	this.seeds = "'新组织可以连接的其他节点监听地址(<IP>:<Port>) 或者 <节点名称>:<Port>，多个以英文逗号分割'|input";
	this.sign_by_CA = "'CA节点对<本节点公钥><新组织名称>的签名结果 或者 <CA节点的名字>?<CA节点私钥密码>'|input";
	this.genesisnode = "'初始成员节点，如果genesisnode中没有要加入节点的信息，这里可不填。格式：<节点名字>?<节点密码>(amount:100,is_ca:true);<节点名字>?<节点密码>'|input"
	this.configfile = "'~/.ann_runtime/config.toml'|file";
}

function OrgLeave() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.privkey = "'节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.orgid = "'要离开的组织名称'|input";
}

function EventUploadCode() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.privkey = "'节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.target = "'接受命令的组织名称，通常是基础组织名称'|input";
	this.code_text = "'事件触发执行的lua code'|text";
	this.ownerid = "'代码所属的组织名称'|input";
}

function EventRequest() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.privkey = "'节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.target = "'接受命令的组织名称，通常是基础组织名称'|input";
	this.source = "'事件触发组织的名称'|input";
	this.source_hash = "'事件要触发的code_hash'|input";
	this.listener = "'监听节点的组织名称'|input";
	this.listener_hash = "'监听节点收到事件触发的code_hash'|input";
}

function EventUnsubscribe() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	this.privkey = "'节点私钥，如果backend填名字，这里填私钥密码'|input";
	this.target = "'接受命令的组织名称，通常是基础组织名称'|input";
	this.listener = "'监听节点的组织名称'|input";
	this.source = "'事件触发组织的名称'|input";
}

function EvmCreate() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	// this.node_privkey = "节点私钥|input";
	this.target = "'接受命令的组织名称(链名)'|input";
	this.privkey = "账户私钥|input";
	this.params = "合约构造函数参数|input";
	this.abi_definition_text = "合约接口描述|text";
	this.code_text = "合约编译后的字节码|text";
}

function EvmCallOrRead() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	// this.node_privkey = "节点私钥|input";
	this.target = "'接受命令的组织名称(链名)'|input";
	this.privkey = "'账户私钥'|input";
	this.contract = "'合约地址'|input";
	this.method = "'evm执行的函数名'|input";
	this.params = "'调用的函数参数'|input";
	this.abi_definition_text = "'合约接口描述'|text";
}

function IkhofiCreate() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	// this.node_privkey = "节点私钥|input";
	this.target = "'接受命令的组织名称(链名)'|input";
	this.privkey = "'账户私钥'|input";
	this.contractid = "'新合约名字'|input";
}

function IkhofiCall() {
	this.backend = "'节点RPC监听地址(<IP>:<Port>) 或者 <节点名字>'|input";
	// this.node_privkey = "节点私钥|input";
	this.target = "'接受命令的组织名称(链名)'|input";
	this.privkey = "'账户私钥'|input";
	this.contractid = "'合约名字'|input";
	this.method = "'ikhofi执行的函数名'|input";
	this.params = "'ikhofi函数执行参数(‘参数1’, ‘参数2’,..)'|input";
}

function CreateCmd(cmd, op) {
	var cmdObj = new Object;
	switch (cmd) {
		case "special":
			switch (op) {
				case "change_validator":
					cmdObj = (new SpecailChangeValidator());
					break;
			}
			break;
		case "sign":
			cmdObj = (new Sign());
			break;
		case "organization":
			switch (op) {
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
			switch (op) {
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
			switch (op) {
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
			switch (op) {
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

function parseLabel(text) {
	return ``;
};

function parseInput(key, placeholder) {
	return `<div class="ant-row ant-form-item">
				<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
					<label for="nodename" title="">
						<span>${key}</span>
					</label>
				</div>
				<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
				<div class="ant-form-item-control">
					<span class="ant-form-item-children">
						<input type="text" value="" id=${key} name=${key} class="ant-input" placeholder=${placeholder}>
					</span>
				</div>
			</div>
		</div>`;
};

function parseRadio(key) {
	return `<div class="ant-row ant-form-item">
				<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
					<label for="nodename" title="">
						<span>${key}</span>
					</label>
				</div>
			<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
				<div class="ant-form-item-control">
					<span class="ant-form-item-children">
					<div class="ant-radio-group">
					<label class="ant-radio-wrapper ant-radio-wrapper-checked" id="radio-label-1">
						<span class="ant-radio ant-radio-checked" id="radio-span-1">
							<input type="radio" class="ant-radio-input" name="${key}" value="true">
							<span class="ant-radio-inner"></span>
						</span>
						<span>true</span>
					</label>
					<label class="ant-radio-wrapper" id="radio-label-2">
						<span class="ant-radio" id="radio-span-2">
							<input type="radio" class="ant-radio-input" name="${key}" value="false">
							<span class="ant-radio-inner"></span>
						</span>
						<span>false</span>
					</label>
					</div>
						</span>
					</div>
				</div>
			</div>`;
}

function parseSelect(key) {
	return `<div class="ant-row ant-form-item">
				<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
					<label for="nodename" title="">
						<span>${key}</span>
					</label>
				</div>
				<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
					<div class="ant-form-item-control">
						<span class="ant-form-item-children">
						
						</span>
					</div>
				</div>
			</div>`;
}

function parseTableStyle(label, inner) {
	return `<div class="ant-row ant-form-item">
	<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
		<label for="nodename" title="">
			<span>${label}</span>
		</label>
	</div>
	<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
		<div class="ant-form-item-control">
			<span class="ant-form-item-children">
			${inner}
			</span>
		</div>
	</div>
</div>`;
}

function parseFile(key, inner) {
	`<div class="ant-row ant-form-item">
				<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
					<label for="nodename" title="">
						<span>${key}</span>
					</label>
				</div>
				<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
					<div class="ant-form-item-control">
						<span class="ant-form-item-children">
						${inner}
						</span>
					</div>
				</div>
			</div>`
}

function acceptFile(file) {
	if (file.indexOf("config") > -1) {
		return ".toml";
	}
	//	if (file.indexOf("genesis") > -1){
	//		return ".json";
	//	}
	return ".json";
}

function innerHTML(type, key, defvalue) {
	var str = "";
	switch (type) {
		case "file":
			str = "<span class='form-group'><input class='ant-input' name='" + key + "_dir' id='" + key + "_dir' style='width:200px'><button style='margin-left:18px' class='ant-btn ant-btn-primary'>上传文件</button><input type='file' name='" + key + "_file' id='" + key + "_file' accept='" + acceptFile(key) + "' class='input-file' onchange='ChangeFile(\"" + key + "\", this.value, this.files)'></span>";
			break;
		case "bool":
			str = "<div class='radio-box'><input type='radio' id='" + key + "_T' name='" + key + "' checked><label for='" + key + "_T'>true</label></div><div class='radio-box'><input type='radio' id='" + key + "_F' name='" + key + "'><label for='" + key + "_F'>false</label></div>";
			break;
		case "text":
			str = "<textarea name='" + key + "' id='" + key + "' rows=4 class='ant-input'></textarea>";
			break;
		case "list":
			// TODO only support appnames_list now, want to fit more types
			str = "<input list='_appnames_list' name='" + key + "' id='" + key + "'> <datalist id='_appnames_list'><option label='evm' value='evm'><option label='ikhofi' value='ikhofi'><option label='noop' value='noop'><option label='remote' value='remote'></datalist>";
			break
		default:
			str = "<input type='text' class='input-text' placeholder='" + defvalue + "' name='" + key + "' id='" + key + "' autocomplete='off'>";
	}
	return str;
}

function parseType(key) {
	if (key.indexOf("file") > -1) {
		return "file";
	}
	if (key.indexOf("_list") > -1) {
		return "list";
	}
	if (key.indexOf("_text") > -1) {
		return "text"
	}
	if (key.indexOf("is") == 0) {
		return "bool";
	}
	return key;
}

function parseType(value) {
	const type = value.split('|')[1];

	switch (type) {
		case 'input':
			return
	}
}

function jointHTML(key, value) {
	//var keytype = parseType(key);
	const keyvalue = value.split('|')[0];
	const keytype = value.split('|')[1];
	switch (keytype) {
		case "file":
			str = parseTableStyle(key + "：", innerHTML(keytype, key, value)) + parseTableStyle("", innerHTML("text", key, value));
			break;
		case "checkbox":
			str = parseRadio(key);
			break;
		case "input":
			str = parseInput(key, keyvalue);
			//str = parseTableStyle(key+"：", innerHTML(keytype, key, value));
			break;
		case "text":
			str = parseTableStyle(key, innerHTML("text", key, value));
			break;
	}
	return str;
}

function AddTable(cmdObj) {
	var html = "";
	// html=jointHTML("backend", "tcp://0.0.0.0:46657");
	html = html + `<input id="cmd" name="cmd" type="hidden"/>
	<input id="op" name="op" type="hidden"/>
	<input id="filediv" name="filediv" type="hidden"/>`
	for (var key in cmdObj) {
		if (typeof (cmdObj[key]) == "function") {
			continue;
		}
		html = html + jointHTML(key, cmdObj[key]);
		continue;
	}
	html = html + `<div class="ant-row ant-form-item">
	<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-xs-offset-0 ant-col-sm-20 ant-col-sm-offset-4">
		<div class="ant-form-item-control">
			<span class="ant-form-item-children">
				<div id="extra_button" name="extra_button"> 
			</span>
		</div>
	</div>
</div>`;
html = html + `<div id="result" class="text-c" name="result"></div>`;
	return html;
}

function ChangeFile(keyid, value, file) {
	document.getElementById(keyid + "_dir").value = value;
	ShowFile(keyid, file[0]);
}

function ShowFile(keyid, file) {
	var reader = new FileReader();
	reader.onload = function (progressEvent) {
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
		for (var i = 0; i < strs.length; i++) {
			theRequest[strs[i].split("=")[0]] = unescape(strs[i].split("=")[1]);
		}
	}
	return theRequest;
}
