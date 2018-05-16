# xlsx2pb

将预定义的xlsx文件转成proto文件和data文件

## 安装

`go get -u github.com/cittie/xlsx2pb`

`go install github.com/cittie/xlsx2pb`

## 使用

在xlsx目录创建 "xlsx_xxxxx.config"形式的配置文件，内容如下：

`SHEETNAME4 XLSXFILENAME2.xlsx`

同样结构的表可以简化成这样：

`SHEETNAME1,SHEETNAME2,SHEETNAME3 XLSXFILENAME1.xlsx`

然后运行如下命令:

`xlsx2pb xlsx_input_path proto_output_path`

## 参数

关闭cache的参数:

`-cache=false`

## 注意

* xlsx里的表名必须用英文，全大写且不重复

## 支持的数据类型

* float(float64)
* int32
* int64
* uint32
* uint64
* sint32
* sint64
* string

## 中文特有的吐槽部分

* 写这个工具主要就是应对游戏数据序列化
* 旧的工具链没有缓存机制，当xlsx文件达到某个数量级的时候，在xlsx里改一行，生成数据就得15分钟
* 即使99%文件完全没有变化，依然要跑一遍打包所有数据的流程
* 所以就写了这么个东西，如果xlsx Hash值没变，跳过；表头用来生成proto文件的Hash值不变，跳过；直接打这个文件data部分
* 大概改动不多的时候可以秒出数据吧
* 第一次写这种基础工具，不知道好不好用
* 反正建议不要直接用在生产环境就对了
* 好了，没了……