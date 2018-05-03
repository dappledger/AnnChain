<!DOCTYPE HTML>
<html>

	<head>
		<meta charset="utf-8">
		<meta name="renderer" content="webkit|ie-comp|ie-stand">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1.0,maximum-scale=1.0,user-scalable=no" />
		<meta http-equiv="Cache-Control" content="no-siteapp" />
		<meta name="keywords" content="anntool cmd">
		<meta name="description" content="anntool operators">
		
		<link rel="stylesheet" type="text/css" href="../static/css/iconfont.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/btn.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/card.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/form.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/grid.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/input.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/label.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/modal.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/select.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/radio.css" />
		<link rel="stylesheet" type="text/css" href="../static/css/main.css" />
		<!-- <link href="static/h-ui/css/H-ui.ie.css" rel="stylesheet" type="text/css" />
		<link rel="stylesheet" type="text/css" href="static/h-ui/css/H-ui.css" /> -->
		<link rel="stylesheet" type="text/css" href="../static/css/nav.css" />
		<title>anntool cmd op</title>
	</head>

	<body ontouchstart>
		<div class="ant-card ant-card-padding-transition">
			<div class="ant-card-body">
				<p class="ant-card-subtitle" id="subtitle"></p>
				<form action="" method="post" class="ant-form ant-form-horizontal" id="dataform" enctype='multipart/form-data'></form>
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
											<span class="ant-confirm-title" id="modal_text">添加成功</span></div>
										<div class="ant-confirm-btns">
											<button type="button" class="ant-btn ant-btn-primary" onclick="closeModal();">
												<span>确认</span></button>
										</div>
									</div>
								</div>
							</div>
						</div>
						<div tabindex="0" style="width: 0px; height: 0px; overflow: hidden;">sentinel</div></div>
				</div>
			</div>
		</div>
	</div>

	<script type="text/javascript" src="static/js/jquery.min.js"></script>
	<script type="text/javascript" src="static/js/nav.js"></script>
	<script type="text/javascript" src="static/js/table.js"></script>
	<script type="text/javascript">
	function closeModal() {
			if ($('#modal_info').css('visibility') !== 'hidden');
			$('#modal_info').css('visibility', 'hidden');
		}
		$(document).ready(function () {
			var Request = new Object();
			Request = GetRequest();
			const cmd = Request['cmd'];
			const op = Request['op'];
			let title = '';
			let subtitle = '';
			document.getElementById("dataform").innerHTML = AddTable(CreateCmd(cmd, op));
			document.getElementById("extra_button").innerHTML ="<button  class='ant-btn ant-btn-primary' onclick='return onSubmit(this.value);' value=''>提交</button>";
			document.getElementById("cmd").value = cmd;
			document.getElementById("op").value = op;
			switch (cmd) {
				case "sign":
					title = 'SIGN';
					subtitle = 'sed2551私钥签名';
					break;
				case "special":
					title = 'CHANGE_VALIDATOR';
					subtitle = '改变（非）validator属性';
					$(".ant-radio-input").change(function() {
						const $selectedvalue = $("input[name='isCA']:checked").val();
                    if ($selectedvalue === "true") {
						$("#radio-label-1").addClass('ant-radio-wrapper-checked');
						$("#radio-label-2").removeClass('ant-radio-wrapper-checked');
						$("#radio-span-1").addClass('ant-radio-checked');
						$("#radio-span-2").removeClass('ant-radio-checked');
                    } else {
						$("#radio-label-2").addClass('ant-radio-wrapper-checked');
						$("#radio-label-1").removeClass('ant-radio-wrapper-checked');
						$("#radio-span-2").addClass('ant-radio-checked');
						$("#radio-span-1").removeClass('ant-radio-checked');
                    }
                });
					break;
				case "organization":
					if (op === 'create') {
						title = 'CREATE';
						subtitle = '组织操作：节点创建组织';
					}
					else if (op === 'join') {
						title = 'JOIN';
						subtitle = '组织操作：节点加入组织';
					}
					else if (op === 'leave') {
						title = 'LEAVE';
						subtitle = '组织操作：节点离开组织';
					}
					break;
				case "event":
					if (op === 'uploadcode') {
						title = 'UPLOAD';
						subtitle = '事件操作：上传code';
					}
					else if (op === 'request') {
						title = 'REQUEST';
						subtitle = '事件操作：通讯请求';
					}
					else if (op === 'unsubscribe') {
						title = 'CANCEL';
						subtitle = '事件操作：取消';
					}
					break;
				case "evm":
					if (op === 'create') {
						title = 'CREATE';
						subtitle = '创建合约';
					}
					else if (op === 'call') {
						title = 'CALL';
						subtitle = '调用合约';
						document.getElementById("extra_button").innerHTML = "<button  class='ant-btn ant-btn-primary' onclick='return onSubmit(this.value);' value='read'>查询</button><button class='ant-btn ant-btn-primary' onclick='return onSubmit(this.value);' value='call'>执行</button>";
					}
					break;
				case "jvm":
					if (op === 'create') {
						title = 'CREATE';
						subtitle = '创建合约';
						document.getElementById("filediv").innerHTML="<inputtype='file' name='file' id='file'/>"
					}
					else if (op === 'call') {
						title = 'CALL';
						subtitle = '调用合约';
						document.getElementById("extra_button").innerHTML = "<button  class='ant-btn ant-btn-primary' onclick='return onSubmit(this.value);' value='read'>查询</button><button class='ant-btn ant-btn-primary' onclick='return onSubmit(this.value);' value='call'>执行</button>";
					}
					break;
			}

			$('#title').html(title);
			$('#subtitle').html(subtitle);
		});
		function showVal() {

		}

		function onSubmit(method) {
			if (method != null && method != "") {
				document.getElementById("op").value = method;
			}
			var form = document.getElementById('dataform');
			var formData = new FormData(form);

			$.ajax({
				type: "POST",
				url: "/cmdlist",
				data: formData,
				processData: false,
				contentType: false,
				success: function (msg) {
					// document.getElementById("result").innerHTML = msg;
					if ($('#modal_info').css('visibility') === 'hidden');
					$('#modal_info').css('visibility', 'visible');
					$('.ant-confirm-title').html(msg);
				}
			});
			return false;
		}
	</script>

</body>

</html>
