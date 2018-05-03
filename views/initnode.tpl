<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="renderer" content="webkit|ie-comp|ie-stand">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1.0,maximum-scale=1.0,user-scalable=no" />
		<meta http-equiv="Cache-Control" content="no-siteapp" />
		<!--[if lt IE 9]>
			<script type="text/javascript" src="lib/html5shiv.js"></script>
			<script type="text/javascript" src="lib/respond.min.js"></script>
		<![endif]-->
		<link rel="stylesheet" type="text/css" href="static/h-ui/css/H-ui.min.css" />
		<link rel="stylesheet" type="text/css" href="static/h-ui/lib/Hui-iconfont/1.0.8/iconfont.min.css" />
    <script type="text/javascript" src="static/h-ui/lib/jquery/1.9.1/jquery.min.js"></script>
    <script type="text/javascript" src="static/js/cmds.js"></script>
    <script type="text/javascript">
	function showVal(){
		var cmd = document.getElementById("cmd").value;
		var op = document.getElementById("op").value; 
		document.getElementById("pheader").innerHTML = cmd + " " + op;
		document.getElementById("formhtml").innerHTML = AddTable(CreateCmd(cmd,op));
	}                                                                                   

	function onSubmit(method){
		document.getElementById("method").value=method;
		$.ajax({
			type: "POST",
			url: "/",
			data : $("form#dataform").serialize(),
			success: function(msg) {
				if (method == "genkey"){
					document.getElementById("keyinfo").innerHTML = msg;
				}
				document.getElementById("result").innerHTML = msg;
			}
		});
		return false;
	}
	</script>
		<style type="text/css">
.ui-sortable .panel-header{ cursor:move}
		</style>
		<title>anntool nodes operations</title>
		<meta name="keywords" content="关键词,5个左右,单个8汉字以内">
		<meta name="description" content="网站描述，字数尽量空制在80个汉字，160个字符以内！">
	</head>
	<body ontouchstart onload="showVal()">
		<div class="panel panel-default mt-20">
			<div class="panel-header clearfix">
				<span class="f-l">nodes init</span>
			</div>
		</div>
	<div class="panel-body">
	<div class="panel-header" id="pheader"></div>	
	    <form action="" method="post" class="form form-horizontal responsive" id="dataform" >
		<input id="keyinfo" name="keyinfo" type="hidden"/>
		<input id="cmd" name="cmd" type="hidden" value="initpeer"/>
		<input id="op" name="op" type="hidden" value=""/>
		<input id="method" name="method" type="hidden" value=""/>
		<div id="formhtml"></div>	
	    </form>
	    <div class="row cl">
		    <div class="col-xs-8 col-xs-offset-3">
		    <button class="btn btn-primary"  onclick="return onSubmit(this.value);" value="genkey">生成公私钥</button>
		    <button class="btn btn-primary"  onclick="return onSubmit(this.value);" value="init">初始化节点</button>
		    <button class="btn btn-primary"  onclick="return onSubmit(this.value);" value="run">运行节点</button>
		    <button class="btn btn-primary"  onclick="return onSubmit(this.value);" value="close">关闭界面服务</button>
		    </div>
	    </div>
		<div id="result" class="text-c" name="result"></div>
	</div>
	</body>
</html>
