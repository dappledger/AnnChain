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
		<style type="text/css">
.ui-sortable .panel-header{ cursor:move}
		</style>
		<title>anntool nodes operations</title>
		<meta name="keywords" content="关键词,5个左右,单个8汉字以内">
		<meta name="description" content="网站描述，字数尽量空制在80个汉字，160个字符以内！">
<style>
        .formControls input {
            border-radius: 4px;
            padding-left: 10px;
            line-height: 30px;
            font-family: '微软雅黑';
        }
        
        .formControls .btn-upload input {
            border-radius: 5px 0 0 5px;
        }
        
        .formControls .btn-upload .btn {
            border-radius: 0 5px 5px 0;
        }
        
        .formControls textarea {
            width: 100%;
            -webkit-border-radius: 5px;
            -moz-border-radius: 5px;
            border-radius: 5px;
            border-color: #ddd;
            min-width: 100%;
            max-width: 100%;
            min-height: 100px;
            padding: 4px 8px;
            font-size: 14px;
            -webkit-box-sizing: border-box;
            box-sizing: border-box;
        }
        
        .btn {
            -webkit-border-radius: 5px;
            -moz-border-radius: 5px;
            border-radius: 5px;
        }
        
        .panel-body>.row {
            margin-top: 20px;
        }
        
        .panel-body>.row .btn {
            margin-right: 10px;
        }
        
        .fade {
            opacity: 1;
            -webkit-transition: opacity .15s linear;
            -o-transition: opacity .15s linear;
            transition: opacity .15s linear;
        }
        
        .fade.in {
            opacity: 0.5
        }
        
        .modal-open {
            overflow: hidden
        }
        
        .modal {
            position: fixed;
            left: 0;
            top: 140px;
            right: 0;
            bottom: 0;
            z-index: 1040;
            display: block;
            overflow: hidden;
            -webkit-overflow-scrolling: touch;
            outline: 0;
        }
        
        .modal.hide {
            display: none;
        }
    </style>
</head>

<body ontouchstart onload="showVal()">
    <div class="panel panel-default mt-20">
        <div class="panel-header clearfix">
            <span class="f-l">nodes init</span>
        </div>
    </div>
    <div class="panel-body">
        <div class="panel-header" id="pheader"></div>
        <form action="" method="post" class="form form-horizontal responsive" id="dataform">
            <input id="keyinfo" name="keyinfo" type="hidden" />
            <input id="cmd" name="cmd" type="hidden" value="initpeer" />
            <input id="op" name="op" type="hidden" value="" />
            <input id="method" name="method" type="hidden" value="" />
            <div id="formhtml"></div>
        </form>
        <div class="row cl">
            <div class="col-xs-8 col-xs-offset-3">
                <button class="btn btn-primary" onclick="return onSubmit(this.value);" value="genkey">生成公私钥</button>
                <button class="btn btn-primary" onclick="return onSubmit(this.value);" value="init">初始化节点</button>
                <button class="btn btn-primary" onclick="return onSubmit(this.value);" value="run">运行节点</button>
                <button class="btn btn-primary" onclick="return onSubmit(this.value);" value="close">关闭界面服务</button>
            </div>
        </div>
        <div id="result" class="text-c" name="result"></div>
        <div id="modal-tip" class="modal hide" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
            <div class="modal-backdrop fade in"></div>
            <div class="modal-dialog">
                <div class="modal-content radius">
                    <div class="modal-header">
                        <h3 class="modal-title">提示</h3>
                        <a class="close" data-dismiss="modal" aria-hidden="true" href=" ">×</a >
                    </div>
                    <div class="modal-body" style="font-size: 16px;">
                        <p>对话框内容…</p >
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-primary guanbi"> 确 定 </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <script type="text/javascript">
        function openModal(txt) {
            $('.modal-body').html(txt)
            $("#modal-tip").removeClass('hide').addClass('open')
        }

        function closeModal() {
            $("#modal-tip").removeClass('open').addClass('hide')
        }
        $(function() {
            var modal = $("#modal-tip")
            $('.close').on('click', function() {
                closeModal()
            })
            $('.guanbi').on('click', function() {
                closeModal()
            })
        })

        function showVal() {
            var cmd = document.getElementById("cmd").value;
            var op = document.getElementById("op").value;
            document.getElementById("pheader").innerHTML = cmd + " " + op;
            document.getElementById("formhtml").innerHTML = AddTable(CreateCmd(cmd, op));
        }

        function onSubmit(method) {
            document.getElementById("method").value = method;
            // openModal(method)
            $.ajax({
                type: "POST",
                url: "/",
                data: $("form#dataform").serialize(),
                success: function(msg) {
                        openModal(msg)
                },
		fail: function(msg){
                        openModal(msg)
		}
            });
            return false;
        }
    </script>
</body>
</html>
