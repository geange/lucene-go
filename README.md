# lucene-go

## 概要

> 兼容 Apache Lucene 8.11.2 的 Go版Lucene

开始这个项目的初衷是因为`Elasticsearch`的菜谱故事过于科幻。在了解了搜索相关的基础知识后，草率开始了项目的开发。

原本我希望实现一个Go版的ES，不过现在距离这个目标还有一点点遥远。
更重要的是需要尽快完善lucene-go项目，使其早日达到一个完全可用的状态。这个可用的状态包含但不限于以下几点：

* 完善代码，原本的代码是将Java翻译成Go，存在很多的不足，下一阶段的目标是让代码更加像gopher写的🐶
* 完善单元测试，单测是保证代码质量的最理想方案～
* 完善用例，现在用例的质量还比较低，仅仅是我个人用于本地测试的一些简单的案例
* 完善文档，这个会在完善代码的过程中一并进行，让使用者可以更容易获取自己想要的内容（毕竟Lucene库的代码量非常大～）

项目的目标在开发的过程中几经波折，遇到的困难远超预期，语言的差异以及原理性知识的缺乏，经过一年的开发逐步完成下面几大模块的开发：

* core/store: lucene的存储模块，主要负责数据的序列化处理
* core/document: 定义lucene的一些搜索相关的数据结构
* core/index: lucene索引的实现，也是对外暴露主要的包
* core/search: 主要包含query的实现（用于在索引中进行数据检索）
* memory: 实现了一个内存实现的搜索引擎，一个简化版的Lucene
* util/fst: FST的实现（Lucene的重要的数据结构）
* util/automaton: 自动机的实现
* codes: 序列化相关，当前仅支持simpleText（使用纯文本信息记录索引数据）的格式

> 需要注意的是，当前项目并不完善！请勿用于任意项目～

## 尝试

### 最新的Go版本

如果需要尝试Lucene-go，需要安装go1.21，因为开发阶段使用到go1.21版本的`cmp`包。

> 在开发阶段对新技术的使用会比较激进，会经常性尝试使用最新的Go技术/标准（例如最新的Go版本、范型语法等）。

安装go1.21rc3的方式如下

```go
go install golang.org/dl/go1.21rc3@latest
go1.21rc3 download
```

### 案例

[更多案例](https://github.com/geange/lucene-go-example)

如何使用`IndexWriter`写入数据

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/store"
)

func main() {
	dir, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}

	codec := simpletext.NewSimpleTextCodec()
	similarity := search.NewCastBM25Similarity()

	config := index.NewIndexWriterConfig(codec, similarity)

	writer, err := index.NewIndexWriter(dir, config)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := writer.Commit(context.Background())
		if err != nil {
			fmt.Println(err)
		}
	}()

	{
		doc := document.NewDocument()
		doc.Add(document.NewStoredFieldAny("a", 74, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a1", 86, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a2", 1237, document.STORED_ONLY))
		docID, err := writer.AddDocument(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(docID)
	}

	{
		doc := document.NewDocument()
		doc.Add(document.NewStoredFieldAny("a", 123, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a1", 123, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a2", 789, document.STORED_ONLY))

		docID, err := writer.AddDocument(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(docID)
	}

	{
		doc := document.NewDocument()
		doc.Add(document.NewStoredFieldAny("a", 741, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a1", 861, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a2", 12137, document.STORED_ONLY))
		docID, err := writer.AddDocument(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(docID)
	}
}

```
