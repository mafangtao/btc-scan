# 错误信息帮助类

返回一个 error ，同时包括自定义的说明及原始的错误信息


    host,err:=os.hostname();
    if err!=nil{
    return goerr.NewError(err,"取主机名时出错")
    }
