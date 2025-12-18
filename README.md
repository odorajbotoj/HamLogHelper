# HamLogHelper

HamLogHelper

by BG4QBF

## 页面说明

+ 左侧
  + 上半部分为已经记录的日志. 提交新日志时会自动滚动到底端
  + 下半部分为正在输入的日志. "锁定" 可以保持当前信息提交时不改变
+ 右侧
  + 上半部分为地图, 可以查看已通联的友台 QTH (基于天地图)
  + 下半部分为搜索框, 可以搜索地点, 点击地点会将信息填入 "对方台址" 一栏

## 使用说明

+ 在可执行文件同级目录下创建 `tianditu-key.txt` , 填入天地图 api-key
+ 访问浏览器 `localhost:5973` 打开前端页面

## 数据说明

+ 标准 csv (使用 `golang: encoding/csv` 库)
+ 前后端除握手测试外, 交换数据均为 json 格式
+ 数据字段如下
  1. 位号 - index
  2. 呼号 - callsign
  3. 日期时间 ( UTC ) - dt
  4. 频率 - band
  5. 模式 - mode
  6. 信号 - rst
  7. 对方设备 - rrig
  8. 对方功率 - rpwr
  9. 对方天线 - rant
  10. 对方台址 - rqth
  11. 己方设备 - trig
  12. 己方功率 - tpwr
  13. 己方天线 - tant
  14. 己方台址 - tqth
  15. 备注 - rmks

## LICENSE

MIT
