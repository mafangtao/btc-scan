##  简介

btc-scan是一个用go语言实现的比特币钱包服务，旨在为比特币钱包客户端提供发起交易，查询UTXO，查询历史交易记录等接口。其基本原理是从全节点同步并解析交易数据，保存到本地leveldb数据库。
详细的设计思路见 [《如何设计一个比特币钱包服务》](https://github.com/liyue201/btc-wallet-service-design)

## 部署

#### 拉取代码
```
git clone https://github.com/liyue201/btc-scan
```

####  编译docker镜像
```
./build_docker.sh
```

#### 使用dcoker-comopse部署
```
docker-compose up -d
```

## 注意事项 
需要等待所有区块交易同步完成之后，才能查询到正确的UTXO。  
- 可以使用 `docker logs -f btcd` 查看btcd同步情况。  
- 使用 `docker logs -f btc_scan` 查看btc-scan同步情况。

## API文档

详见 [《API文档》](/docs/api.md)