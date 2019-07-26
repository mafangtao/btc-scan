# API文档

- [文档概述](#introduction)
    - [HTTP 相关](#http)
    - [状态码](#statusCode)
    - [内部错误状态码](#innerStatusCode)
- [接口列表](#api)
    - [获取交易记录](#txs)
    - [获取utxo](#utxo)
    - [发送交易](#sendtx)
    - [获取最新区块高度](#bestheith)


## <a name="introduction">文档概述</a>

### <a name="http">HTTP 相关</a>

请求数据:

- POST:
- 请求体: `json`
- 请求首部:
	- `json` : `Content-type: application/json; charset=utf-8`
响应数据:

- 响应首部:
	- `Content-type: application/json; charset=utf-8`
- 响应体:

```json
{
  "data": {
  },
  "code":40002,
  "message":"请求参数不正确"
}
```

| 字段名        | 类型          | 描述                                |
| ---------- | ----------- | --------------------------------- |
| data       | Json object | 接口返回的数据, 不同接口返回不同数据, 如无数据则为空 `{}` |
| message    | string      | 请求响应消息, 可以直接展示给用户

### <a name="statusCode">状态码</a>
接口的请求状态可以通过 HTTP 的状态码简单判断, 下面列出一些通用的状态码:

| 错误码  | 描述                    |
| ---- | --------------------- |
| 200  | 请求成功, 返回正确的数据         |
| 400  | 参数错误, 可能缺少参数或者类型不正确   |
| 401  | 用户认证信息失效, 可能已经过期或者不正确 |
| 403  | 无相关操作权限               |
| 404  | 资源未找到                 |
| 408  | 请求超时                  |
| 500  | 系统内部错误                |

约定规则:

- 凡是非 2xx 的状态码均为请求失败
- 如果返回 401 状态码需要引导用户重新登录

### <a name="innerStatusCode">内部错误状态码</a>

接口的请求状态可以通过 HTTP 返回 Json 中的 code 状态码简单判断, 下面列出错误状态码:

| 错误码   | 描述               |
| ----- | ---------------- |
| 0  | 成功	|
| 40002 | 参数不合法，请检查参数	|
| 50000 | 服务器内部错误 |

## <a name="api">接口列表</a>
### <a name="txs">获取交易记录</a>
- 接口功能: 获取交易记录
- 请求地址: `/api/v1/txs?address=mxzVuHu1xg5Da4H79maoLsoH665B4pKHj8&limit=10&order=-1&prevminkey=&prevmaxkey`
- 请求方式: `GET`
- 请求参数:

| 参数名            | 类型         | 描述                | 必须   | 举例             |
| -------------- | ---------- | ----------------- | ---- | ----------------- |
| address| string  |  比特币地址| 是 |
| limit| int  |  返回条数上限，默认10| 否 |
| order| int  | -1：向后翻页， 1：向前翻页，默认-1| 否|
| prevminkey| string  |  上一次请求返回的minkey，用于翻页| 否 |
| prevmaxkey| string  |  上一次请求返回的maxkey，用于翻页| 否 |

- 返回数据:

	```json
	{
	    "data": {
	       "minkey":"",
	       "maxkey":"",
	       "txs":[{
	            "txId":"",
	            "blockTime": 0,
	            "blockHeight": 1100,
	            "blockHash":"",
	            "inputs":[{
	                 "txId":"",
	                 "vout":1,
	                 "addr":"",
	                 "amount":100
	            }],
	            "outputs":[{
	                 "vout":1,
	                 "addr":"",
	                 "amount":100,
	                 "scriptPubKey":""
	            }]
	       }]
	    },
	    "message": "Success",
	    "code": 0
	}
	```

- 返回字段:

    | 字段名        | 类型          | 描述                |
    | ---------- | ----------- | ----------------- |
    | data       | json object | 返回的数据             |
    | message    | string      | 请求响应消息, 可以直接展示给用户 |
    | code       | int      | 请求返回错误码           |

- `data` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | minkey | string |  |
    | maxkey | 最为翻页查询的参数|  |
    | txs| json array | 交易数组  |

- `txs` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | txId| string | 交易id|
    | blockTime| int64 |区块时间戳 |
    | blockHeight| int64 | 区块高度|
    | blockHash| string |区块hash，空表示未上链|
    | inputs| array | 交易输入|
    | outputs| array |交易输出 |

- `inputs` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | txId| string | 交易id|
    | vout| int |前一个交易输出的下标  |
    | addr| string |比特币地址|
    | amount| int64 | 金额|

- `outputs` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | vout| int |输出的下标  |
    | addr| string |比特币地址|
    | amount| int64 | 金额|
    | scriptPubKey|string| 公钥脚本|

### <a name="utxo">获取utxo</a>

- 接口功能: 获取utxo
- 请求地址: `/api/v1/addr/{addr}/utxo`
- 请求方式: `GET`

- 返回数据:

	```json
	{
	    "data":{
    	    "utxos":[{
    	        "txId": "",
    	        "vout": 1,
    	        "amount":10000000,
    	        "scriptPubKey":""
    	    }]
	    },
	    "message": "Success",
	    "code": 0
	}
	```

- 返回字段:

    | 字段名        | 类型          | 描述                |
    | ---------- | ----------- | ----------------- |
    | data       | json object | 返回的数据             |
    | message    | string      | 请求响应消息, 可以直接展示给用户 |
    | code       | int      | 请求返回错误码           |

- `data` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | utxos | array | 数组 |

- `utxos` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | txId | string | 交易id |
    | vout | int | 输出下标|
    | amount| int64 |金额 |
    | scriptPubKey| string |  公钥脚本  |


### <a name="sendtx"> 发送交易</a>

- 接口功能: 发送交易
- 请求地址: `/api/v1/tx/send`
- 请求方式: `POST`

- 请求参数:
rawtx=0100000002D23658D4E039799FFE6C3EC1290C3A0DB8F7A21EAB77C272317843F61EDB3A81000000006B483045022100AC7AE9A9229C81BF68539826435648899BF0213E352680947709379B693BA43D0220718FDC924EAF90B9714D310F5E1D0B1CCD48C2263C0E2738F81E0CFA3F172A11012103374330BB1A7CD1366807576FE84E5142E8A75D0D0141AAE4B50D450724E9FF07FFFFFFFFD23658D4E039799FFE6C3EC1290C3A0DB8F7A21EAB77C272317843F61EDB3A81010000006B4830450221008CF4A64584AA4D534470CE3CE92F063474104AFAC3E9B3436C1EE1E94E999D78022071C501DDA0DC8D1703C2616D6DFAE529376359CB773F2DB05AF320C106A9AFC3012103374330BB1A7CD1366807576FE84E5142E8A75D0D0141AAE4B50D450724E9FF07FFFFFFFF02E8030000000000001976A9144CEE49C798D67431B084467CAB90400834ED217B88ACF12E0000000000001976A9144CEE49C798D67431B084467CAB90400834ED217B88AC00000000


    | 参数名            | 类型         | 描述                | 必须   | 举例                |
    | -------------- | ---------- | ----------------- | ---- | ----------------- |
    | rawtx| string  |  已经签名的交易数据| 是 |

- 返回数据:

	```json
	{
	    "data": {
	        "txid":""
	    },
	    "message": "Success",
	    "code": 0
	}
	```

- 返回字段:

    | 字段名        | 类型          | 描述                |
    | ---------- | ----------- | ----------------- |
    | data       | json object | 返回的数据             |
    | message    | string      | 请求响应消息, 可以直接展示给用户 |
    | code       | int      | 请求返回错误码           |

- `data` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | txid | string | 交易id |


### <a name="bestheith">获取最新区块高度</a>

- 接口功能: 获取最新区块高度
- 请求地址: `/api/v1/best_height`
- 请求方式: `GET`

- 返回数据:

	```json
	{
	    "data": {
	        "bestHeight": 100000
	    },
	    "message": "Success",
	    "code": 0
	}
	```

- 返回字段:

    | 字段名        | 类型          | 描述                |
    | ---------- | ----------- | ----------------- |
    | data       | json object | 返回的数据             |
    | message    | string      | 请求响应消息, 可以直接展示给用户 |
    | code       | int      | 请求返回错误码           |

- `data` 字段:

    | 字段名      | 类型     | 描述       |
    | -------- | ------ | -------- |
    | bestHeight | int64 | 区块高度 |


