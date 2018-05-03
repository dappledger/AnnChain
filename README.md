full-feature annchain

# IKHOFI README

## Set JAVA_HOME

    wget https://cfapi-test.zhongan.io/download/jdk-8u101-linux-x64.tar.gz
    tar zxvf jdk-8u101-linux-x64.tar.gz

Edit profile:

    # for java
    export JAVA_HOME=$HOME/jdk-8u101-linux-x64
    export JRE_HOME=${JAVA_HOME}/jre
    export CLASSPATH=.:${JAVA_HOME}/lib:${JRE_HOME}/lib
    export PATH=${JAVA_HOME}/bin:${JRE_HOME}/bin:$PATH

## Copy ikhofi to JAVA_HOME

    wget https://cfapi-test.zhongan.io/download/ikhofi-all-0.6.5-jar-with-dependencies.jar
    cp -rf ikhofi-all-0.6.5-jar-with-dependencies.jar $JAVA_HOME/jre/lib/ext/

## Run ikhofi server

    wget https://cfapi-test.zhongan.io/download/ikhofi-api-0.6.5.jar
    java -jar target/ikhofi-server-0.6.5.jar --server.port=8080

## Create ikhofi org

Need to add ikhofi_addr to ikhofi org config.toml:

    # This is a TOML config file.
    # For more information, see https://github.com/toml-lang/toml
    
    seeds = ""
    appname = "ikhofi"
    p2p_laddr = "0.0.0.0:50011"
    log_path = "/Users/shilei/ann/ikhofi/"
    auth_by_ca = false
    signbyCA = ""
    cosi_laddr = "0.0.0.0:55011"
    ikhofi_addr = "http://127.0.0.1:8080"
    
And then create ikhofi org:

    ./build/anntool --callmode="commit" --backend="tcp://127.0.0.1:46657" --target="annchain-civil" organization create --genesisfile genesis.json --configfile config.toml --privkey 6EBFC5A9B535A80D5BDEC0141502193F748FA6A9A02E1B1E9C59B010DA36BCBBA3C7622F824209127D7DFF74B8EAA48B169D97E015F358A46EA896C297244FE9

## Create a java contract

    ./build/anntool --target="ikhofi" --backend="tcp://127.0.0.1:46657" ikhofi execute -contractid system -method "deploy('StorageContract01', './scripts/examples/contracts/StorageContract.class')" -privkey abee332ae1f38479d16c709fdcb601d2b07dceab8fdd2d8c932f0ed01a988264
    
## Query contract id

    ./build/anntool --target="ikhofi" --backend="tcp://127.0.0.1:46657" ikhofi query -contractid system -method "contract('StorageContract0')" -privkey abee332ae1f38479d16c709fdcb601d2b07dceab8fdd2d8c932f0ed01a988264
    
## Execute contract

    ./build/anntool --target="ikhofi" --backend="tcp://127.0.0.1:46657" ikhofi execute -contractid StorageContract01 -method "setData('a', '1')" -privkey abee332ae1f38479d16c709fdcb601d2b07dceab8fdd2d8c932f0ed01a988264
    
## Query contract

    ./build/anntool --target="ikhofi" --backend="tcp://127.0.0.1:46657" ikhofi query -contractid StorageContract01 -method "getData()" -privkey abee332ae1f38479d16c709fdcb601d2b07dceab8fdd2d8c932f0ed01a988264
