function InitPeer(){
	this.config_path= "/配置/存储/路径";
	this.chainid = "链名";
	this.peer_privkey = "";
	this.p2p_port = "46656";
	this.rpc_port = "46657";
	this.event_port = "46650";
	this.peers = "<IP1>:<Port1>,<IP2>:<Port2>...";
	this.auth_privkey = "CA签名节点私钥";
	this.log_path = "/日志/路径";
	this.genesisfile = "";
}

function CreateCmd(cmd , op){
	var cmdObj = new Object;
	switch (cmd){
		case "initpeer":
			cmdObj = (new InitPeer());
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
