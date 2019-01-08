# 一键安装annChain.Genesis脚本使用说明

本文档能够让初学者快速体验annChain.Genesis链。提供了annChain.Genesis的快速安装和annChain.Genesis节点的快速部署步骤。

## 注意事项：

- 安装环境需满足annChain.Genesis运行的配置需求[操作手册1.1节](https://github.com/dappledger/AnnChain/tree/master/doc/manual)；
- genesis_install.sh在CentOS和Ubuntu版本测试成功；
- 如遇到执行genesis_install.sh中下载失败，一般和网络状况有关，请删除掉所有git clone下载的内容，重新执行genesis_install.sh.
- 某些步骤会提示输入root密码。

## Genesis单节点部署

1. 下载脚本

   wget https://github.com/dappledger/AnnChain/blob/master/scripts/genesis_install.sh

   wget https://github.com/dappledger/AnnChain/blob/master/scripts/genesis_service.sh

2. 执行脚本

   chmod 755 genesis_install.sh genesis_service.sh

   ./genesis_install.sh

   ./genesis_service.sh

   ```
   Please enter the Chain-ID:test
   Please enter the P2P Listen on port:80
   Please enter the RPC Listen on port:81
   Please enter the SEEDS P2P NODE format(IP:PORT):127.0.0.1:80
   
                _                  ____            _     _
               / \    |\    |\    |    \|    |    / \    |\    |
              /   \   | \   | \   |     |    |   /   \   | \   |
             / ___ \  |  \  |  \  |     |____|  /_____\  |  \  |
            /       \ |   \ |   \ |     |    | /       \ |   \ |
           /         \|    \|    \|____/|    |/         \|    \|
   
   
           Genesis has been successfully built. 00:11:19
   
           To verify your installation run the following commands:
           For more information:
           AnnChain website: www.annchain.io/#/
           AnnChain Telegram channel @ www.annchain.io/#/news
           AnnChain resources: https://github.com/dappledger/AnnChain
           AnnChain Stack Exchange: https://...
           AnnChain wiki: https://github.com/dappledger/AnnChain/blob/master/README.md
   ```
