# HamLogHelper

HamLogHelper

by BG4QBF

v1.4.0

## 使用视频

[[业余无线电] 点名记录助手 版本更新介绍 v1.3.0 指定端口 添加模板 快速选中 - Bilibili](https://www.bilibili.com/video/BV1GRBZBLE6c/)

## 页面截图

+ 主页
  ![mainpage](https://raw.githubusercontent.com/odorajbotoj/HamLogHelper/refs/heads/main/assets/mainpage.png)
+ 主页 (无地图)
  ![mainpage-nomap](https://raw.githubusercontent.com/odorajbotoj/HamLogHelper/refs/heads/main/assets/mainpage-nomap.png)
+ 导出
  ![export](https://raw.githubusercontent.com/odorajbotoj/HamLogHelper/refs/heads/main/assets/export.png)
+ 数据库编辑
  ![dbeditor](https://raw.githubusercontent.com/odorajbotoj/HamLogHelper/refs/heads/main/assets/dbeditor.png)

## 页面说明

+ 上半部分
  + 左侧为已经记录的日志. 提交新日志时会自动滚动到底端
  + 右侧为地图, 可以查看已通联的友台 QTH (基于天地图, 无 api-key 时不显示)
+ 下半部分
  + 左侧为正在输入的日志. "锁定" 可以保持当前信息提交时不改变
  + 右侧为搜索框, 可以根据模板搜索呼号或根据字典搜索模式 / 设备 / 天线 / 功率, 以及根据字典和天地图 API 搜索 QTH (若有 api-key), 点击地点会将信息填入相应栏

## 快速上手

1. 在可执行文件同级目录下创建 `tianditu-key.txt` , 填入天地图 api-key (若无 api-key , 将自动禁用天地图相关功能)
2. 根据个人需要创建模板与字典
3. 运行主程序. 默认会绑定本机 5973 端口, 您可以使用命令行参数 `-a` 来指定绑定的端口, 如 `-a 0.0.0.0:1234` 或 `-a :7359`.
   + **警告: 您不应将服务暴露在公网下! 注意本程序无多用户隔离和安全防护! 允许自定绑定端口只是为了使用方便!**
4. 在浏览器中访问您设置的地址以打开前端页面
5. 输入名称以打开日志, 程序会保存一份 `hjl` 文件 (`HamJsonLogs`)

## 填写注意

+ 模式仅允许搜索提示区列出的模式
+ 频率输入格式为 `%f/%f` , 前后两个小数, 中间斜杠隔开, 代表发射频率与频差. 如 `430.610/+9.0` , `438.5/+0.0`, `439.645/-9.0`

## 功能详情

### 数据编辑

#### 录入

+ 按照字段输入信息.
+ 呼号自动转大写, 并显示模板信息 (如有).
+ 对方设备 / 天线 / 功率 / 台址在获得焦点时显示输入历史 (如有), 输入框内容更新时显示字典补全 (如有).
  + 你可以使用 `Ctrl+<数字>` 或 `Meta+<数字>` (Windows 上 `Meta` 为 `Windows 徽标` 键, Mac 上 `Meta` 为 `Command` 键) 来快速选择右侧的输入建议, 譬如, 当你按下 `Ctrl+1` 会自动填充右侧输入建议的第一项. 数字可取 `1` ~ `9` .
+ 若有天地图 api-key , 台址输入框内容更新时还会额外显示天地图地点搜索建议.
+ 点击记录的位号重新编辑记录.
+ 所有日志变更将在文件重新被打开时生效.

#### 删除

+ 点击 "删除" 会将一行数据标记为删除, 但不会立刻被清理.
+ 标记删除的数据不会影响当前的日志计数.
+ 再次点击位号并提交编辑可以取消删除操作.
+ 所有日志变更将在文件重新被打开时生效.

### 数据导出

当点击 "记录" 旁边的 "导出" 时, 您将打开一个新的页面. 您可以在这里导出您的记录.

若勾选 "仅保存北京时间hh:mm", 则时间将不会保留 UTC 的 `yyyy-mm-ddThh:mm`, 而是仅保留北京时间的小时和分钟 `hh:mm`.

您需要勾选想要导出的列. 导出 CSV 或 XLSX 时, 列的顺序由您勾选的顺序决定 (页面中会有 `导出字段顺序` 作提示).

1. 简单的 ADIF 文件
   + 仅包含如下字段
     1. CALL
     2. BAND
     3. MODE
     4. QSO_DATE
     5. TIME_ON
     6. FREQ
     7. BAND_RX
     8. FREQ_RX
2. 标准 csv (使用 `golang: encoding/csv` 库)
   + 可以自选要导出哪些信息
     1. 位号 - index
     2. 呼号 - callsign
     3. 时间 - dt
     4. 频率 - band
     5. 模式 - mode
     6. 信号 - rst
     7. 对方设备 - rrig
     8. 对方天线 - rant
     9. 对方功率 - rpwr
     10. 对方台址 - rqth
     11. 己方设备 - trig
     12. 己方天线 - tant
     13. 己方功率 - tpwr
     14. 己方台址 - tqth
     15. 备注 - rmks
3. XLSX 文件 (使用 `golang: github.com/xuri/excelize/v2` 库)
   + 可以自选要导出哪些信息
     1. 位号 - index
     2. 呼号 - callsign
     3. 时间 - dt
     4. 频率 - band
     5. 模式 - mode
     6. 信号 - rst
     7. 对方设备 - rrig
     8. 对方天线 - rant
     9. 对方功率 - rpwr
     10. 对方台址 - rqth
     11. 己方设备 - trig
     12. 己方天线 - tant
     13. 己方功率 - tpwr
     14. 己方台址 - tqth
     15. 备注 - rmks

### 模板数据说明

+ json格式, 须命名为 `tmpl.json` 且和可执行文件在同一目录下才可被加载
+ 可以使用自带的工具进行编辑. 出现相同项时, 将自动去重

示例

```json
[
    {
        "callsign": "BG4QBF",
        "rig": "宝锋5RH",
        "ant": "老鹰770拉杆",
        "pwr": "高",
        "qth": "江苏镇江京口区"
    },
    {
        "callsign": "BG4QBF",
        "rig": "泉盛UV-K6",
        "ant": "原装",
        "pwr": "5W",
        "qth": "江苏南京紫金山头陀岭"
    }
]
```

### 字典数据说明

+ json格式, 须命名为 `dict.json` 且和可执行文件在同一目录下才可被加载
+ 可以使用自带的工具进行编辑. 当为空的补全条目失去焦点时, 将被删除. 出现同名项时，将自动进行归并

示例

```json
{
    "rig":{
        "bf": [
            "宝锋5RH",
            "宝锋UV5R",
            "宝锋UV5RMini",
            "宝锋UV32"
        ],
        "b5r": [
            "宝锋5RH",
            "宝锋UV5R",
            "宝锋UV5RMini"
        ],
        "b32": [
            "宝锋UV32"
        ]
    },
    "ant":{
        "2e": [
            "2单元八木"
        ],
        "3e": [
            "3单元八木",
            "3单元CJU"
        ]
    },
    "pwr":{
        "h": [
            "高"
        ],
        "m": [
            "中",
            "满"
        ],
        "max": [
            "满"
        ],
        "l": [
            "低"
        ]
    },
    "qth":{
        "xlw": [
            "江苏南京孝林卫"
        ],
        "yfq": [
            "江苏南京油坊桥"
        ],
        "yht": [
            "江苏南京雨花台"
        ],
        "cpu": [
            "江苏南京中国药科大学"
        ]
    }
}
```

## 鸣谢

+ 江苏省无线电和定向运动协会 BY4RSA

## LICENSE

MIT
