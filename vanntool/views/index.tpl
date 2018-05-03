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
	<div class="head">
		<span>vanntool</span>
	</div>
	<div class="nav">
		<ul>
			<li class="nav-item">
				<a href="javascript:;">
					<i class="iconfont nav-icon icon-nodeinfo"></i>
					<span class="item">节点信息</span>
					<i class="iconfont nav-more icon-right-arrow"></i>
				</a>
				<ul>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="nodesop" target="main">register</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="nodeinfo" target="main">info</a>
						</a>
					</li>
				</ul>
			</li>
			<li class="nav-item">
				<a href="javascript:;">
					<i class="iconfont nav-icon icon-special"></i>
					<span class="item">节点管理</span>
					<i class="iconfont nav-more icon-right-arrow"></i>
				</a>
				<ul>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=special&op=change_validator" target="main">change_validator</a>
						</a>
					</li>
				</ul>
			</li>
			<li class="nav-item">
				<a href="javascript:;">
					<i class="iconfont nav-icon icon-organization"></i>
					<span class="item">子链管理</span>
					<i class="iconfont nav-more icon-right-arrow"></i>
				</a>
				<ul>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=organization&op=create" target="main">create</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=organization&op=join" target="main">join</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=organization&op=leave" target="main">leave</a>
						</a>
					</li>
				</ul>
			</li>
			<li class="nav-item">
				<a href="javascript:;">
					<i class="iconfont nav-icon icon-event"></i>
					<span class="item">事件相关</span>
					<i class="iconfont nav-more icon-right-arrow"></i>
				</a>
				<ul>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=event&op=uploadcode" target="main">uploadcode</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=event&op=request" target="main">request</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=event&op=unsubscribe" target="main">unsubscribe</a>
						</a>
					</li>
				</ul>
			</li>
			<li class="nav-item">
				<a href="javascript:;">
					<i class="iconfont nav-icon icon-evm"></i>
					<span class="item">EVM合约</span>
					<i class="iconfont nav-more icon-right-arrow"></i>
				</a>
				<ul>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" href="cmdlist?cmd=evm&op=create" target="main">create</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" class="item-child" href="cmdlist?cmd=evm&op=call" target="main">call</a>
						</a>
					</li>
				</ul>
			</li>
			<li class="nav-item">
				<a href="javascript:;">
					<i class="iconfont nav-icon icon-evm"></i>
					<span class="item">JVM合约</span>
					<i class="iconfont nav-more icon-right-arrow"></i>
				</a>
				<ul>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" class="item-child" href="cmdlist?cmd=jvm&op=create" target="main">create</a>
						</a>
					</li>
					<li>
						<a href="javascript:;">
							<i class="iconfont nav-icon-dot icon-dot"></i>
							<a class="item-child" class="item-child" href="cmdlist?cmd=jvm&op=call" target="main">call</a>
						</a>
					</li>
				</ul>
			</li>
		</ul>
	</div>
	
    <iframe id="main" name="main" height="1200" width="1000" frameborder="0"></iframe>

    <script type="text/javascript" src="static/js/jquery.min.js"></script>
	<script type="text/javascript" src="static/js/nav.js"></script>
	<script type="text/javascript">
	function closeModal() {
			if ($('#modal_info').css('visibility') !== 'hidden');
			$('#modal_info').css('visibility', 'hidden');
		}
		function onSubmit(method){
			$.ajax({
				type: "POST",
				url: "/nodeinfo",
				data : $("form#dataform").serialize(),
				success: function(msg) {
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