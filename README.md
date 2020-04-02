# 100GB url 文件在有限内存下进行TopN操作
get top n from big url file

## usage
./url100g -s [seperate count] -b [big file fullpath] -o [output dir] -n [top n]
<br>
将100Gurl文件[-b]按行读取，hash之后分为[-s]片，输出到[-o]目录，求TopN[-n]
<br>

### for example
```
./url100g -s 1013 -b "/data/home/twl/gotest/src/github.com/codertwl/url100g/tmp/urls.txt" -o "/data/home/twl/gotest/src/github.com/codertwl/url100g/tmp" -n 100
```

## 算法简介
100GB url 文件在有限内存中进行TopN操作，需要对文件进行分片，满足内存要求后再进行处理
1. 按行读入url,hash 之后存入对应分片文件,默认存入./tmp/seps目录.hash的作用是保证同一url保存到同一文件分片内,hash函数要保证有较低的冲突概率.调整[-s]文件分片参数以满足内存要求
2. 遍历各分片内容，对相同url进行数量统计，统计完成后将同一分片内总数量在前N位的url和其数量按数量由大到小写入对应的新文件中，新文件默认在./tmp/sorts目录中
3. 循环从上一步保存的新文件中每次各取一行，以url统计数量为比较值建立小顶堆，堆大小为N，超过N是进行Pop操作，直到所有文件内容取完一遍，堆中内容即所求结果，排序输出
<br>

## 内存分析
程序运行过程为 (1)从大文件每次读入一行-->(2)计算hash值后存入对应分片-->(3)上一步完成后从对应分片读取url并计数-->(4)写入新文件TopN-->(5)利用小根堆取TopN
<br>
* 其中1，每次操作一行，内存使用可以忽略
* 2,4为输出操作可以忽略
* 5维护一个N项小根堆，内存总量为(N*单项大小),本题N为100，可以忽略
* 3是唯一有可能超出内存限制的步骤，当大量url经过hash后发生碰撞，这些url会被存入同一分片，导致第3步内存被撑爆.
理想情况下hash足够均衡，每个分片获得的不同url的数量都是相等的，则内存占用为 O(totaldiffurls/sepscount).
如果出现严重偏斜，解决办法是使用新hash函数将这类文件再分片，直到不存在此类情况为止
<br>

## 测试数据生成
1. 简单爬取url,使用工具[fetchurl](https://github.com/codertwl/fetchurl)
2. 利用爬取到的url生成随机分布的url,使用工具[randline](https://github.com/codertwl/randline)
