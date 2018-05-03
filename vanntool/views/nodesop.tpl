<!DOCTYPE HTML>
<html>

<head>
	<meta charset="utf-8">
	<meta name="renderer" content="webkit|ie-comp|ie-stand">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1.0,maximum-scale=1.0,user-scalable=no" />
	<meta http-equiv="Cache-Control" content="no-siteapp" />
	<title>anntool cmd list</title>
	<link rel="stylesheet" type="text/css" href="../static/css/main.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/nav.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/iconfont.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/btn.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/card.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/form.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/modal.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/grid.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/input.css" />
	<link rel="stylesheet" type="text/css" href="../static/css/label.css" />
</head>

<body ontouchstart>
	<div class="ant-card ant-card-padding-transition">
		<div class="ant-card-body">
			<p class="ant-card-subtitle">将节点信息记录到服务器，操作节点时可以使用节点名称</p>
			<form action="" method="post" id="dataform" class="ant-form ant-form-horizontal">
				<input id="encnodepk" name="encnodepk" type="hidden" />
				<input id="method" name="method" type="hidden" />
				<div class="ant-row ant-form-item">
					<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
						<label for="nodename" title="">
							<span>节点名称</span>
						</label>
					</div>
					<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
						<div class="ant-form-item-control">
							<span class="ant-form-item-children">
								<input type="text" value="" id="nodename" name="nodename" class="ant-input" placeholder="可使用数字+字母+下划线">
							</span>
						</div>
					</div>
				</div>
				<div class="ant-row ant-form-item">
					<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
						<label for="nodeip" title="">
							<span>节点PRC</span>
						</label>
					</div>
					<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
						<div class="ant-form-item-control">
							<span class="ant-form-item-children">
								<input type="text" value="" id="nodeip" name="nodeip" class="ant-input">
							</span>
						</div>
					</div>
				</div>
				<div class="ant-row ant-form-item">
					<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
						<label for="nodepk" title="">
							<span>节点私钥</span>
						</label>
					</div>
					<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
						<div class="ant-form-item-control">
							<span class="ant-form-item-children">
								<input type="text" value="" id="nodepk" name="nodepk" class="ant-input">
							</span>
						</div>
					</div>
				</div>
				<div class="ant-row ant-form-item">
					<div class="ant-form-item-label ant-col-xs-24 ant-col-sm-4">
						<label for="nodepkpwd" title="">
							<span>私钥密码</span>
						</label>
					</div>
					<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-sm-20">
						<div class="ant-form-item-control">
							<span class="ant-form-item-children">
								<input type="text" value="" id="nodepkpwd" name="nodepkpwd" class="ant-input">
							</span>
						</div>
					</div>
				</div>
				<div class="ant-row ant-form-item">
					<div class="ant-form-item-control-wrapper ant-col-xs-24 ant-col-xs-offset-0 ant-col-sm-20 ant-col-sm-offset-4">
						<div class="ant-form-item-control">
							<span class="ant-form-item-children">
								<button type="submit" class="ant-btn ant-btn-primary" onclick="return onSubmit(this.value);" value="add">
									<span>创建</span>
								</button>
								<button type="submit" class="ant-btn ant-btn-primary" onclick="return onSubmit(this.value);" value="modify">
									<span>修改</span>
								</button>
								<button type="submit" class="ant-btn ant-btn-primary" onclick="return onSubmit(this.value);" value="delete">
									<span>删除</span>
								</button>
							</span>
						</div>
					</div>
				</div>
				<div id="result" class="text-c" name="result"></div>
				<div id="modal_info" style="visibility:hidden">
					<div class="ant-modal-mask"></div>
					<div tabindex="-1" class="ant-modal-wrap " role="dialog">
						<div role="document" class="ant-modal ant-confirm ant-confirm-info" style="width: 416px; transform-origin: 389.5px 204px 0px;">
							<div class="ant-modal-content">
								<button aria-label="Close" class="ant-modal-close">
									<span class="ant-modal-close-x"></span>
								</button>
								<div class="ant-modal-body">
									<div class="ant-confirm-body-wrapper">
										<div class="ant-confirm-body">
											<i class="anticon iconfont anticon-info-circle"></i>
											<span class="ant-confirm-title" id="modal_text">添加成功</span>
										</div>
										<div class="ant-confirm-btns">
											<button type="button" class="ant-btn ant-btn-primary" onclick="closeModal();">
												<span>确认</span>
											</button>
										</div>
									</div>
								</div>
							</div>
							<div tabindex="0" style="width: 0px; height: 0px; overflow: hidden;">sentinel</div>
						</div>
					</div>
				</div>
			</form>
		</div>
	</div>

	<table border="2" width="740">
		<tr>
			<th>index</th>
			<th>name</th>
			<th>rpc</th>
		</tr>
		{{range $key, $val := .nodes}}
		<tr>
			<td>{{$key}}</td>
			<td>{{$val.Name}}</td>
			<td>{{$val.RPCAddr}}</td>
		</tr>
		{{end}}
	</table>

	<script type="text/javascript" src="static/js/jquery.min.js"></script>
	<script type="text/javascript" src="static/js/nav.js"></script>
	<script type="text/javascript" src="static/crypto/rollups/aes.js"></script>
	<script type="text/javascript" src="static/js/crypto.js"></script>
	<script type="text/javascript">
		function closeModal() {
			if ($("#modal_info").css("visibility") !== "hidden");
			$("#modal_info").css("visibility", "hidden");
		}
		function onSubmit(method) {
			const pk = $("#nodepk").val();
			const pwd = $("#nodepkpwd").val();
			$("#encnodepk").val(AESEncrypto(pk, pwd));
			$("#method").val(method);
			$.ajax({
				type: "POST",
				url: "/nodesop",
				data: $("form#dataform").serialize(),
				success: function (msg) {
					//document.getElementById("result").innerHTML = msg;
					if ($("#modal_info").css("visibility") === "hidden");
					$("#modal_info").css("visibility", "visible");
					$(".ant-confirm-title").html(msg);
				}
			});
			return false;
		}
	</script>
</body>

</html>