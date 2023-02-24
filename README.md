# simple-dou

## 极简版抖音

具体功能内容参考飞书说明文档

工程无其他依赖，直接编译运行即可

```shell
go build github.com/Ljkkun/GreenBeanMiners/main
```

### 功能说明

接口功能完善

* 用户登录数据保存在内存中，运行过程中有效
* 视频上传后会保存到本地 public 目录中，访问时用 localhost:8080/static/video_name 即可

### 测试

test 目录下为不同场景的功能测试case，可用于验证功能实现正确性

其中 common.go 中的 _serverAddr_ 为服务部署的地址，默认为本机地址，可以根据实际情况修改
