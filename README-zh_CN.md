# lucene-go

[![GoDoc](https://godoc.org/github.com/geange/lucene-go?status.svg)](https://godoc.org/github.com/geange/lucene-go)
[![Go](https://github.com/geange/lucene-go/actions/workflows/go.yml/badge.svg)](https://github.com/geange/lucene-go/actions/workflows/go.yml)
[![CodeQL](https://github.com/geange/lucene-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/geange/lucene-go/actions/workflows/codeql.yml)

**[English](README.md)**

## 概要

Lucene是一个搜索引擎库。`lucene-go` 是它的Golang版本实现。

### 当前版本

* 仅支持Go1.21+
* 基于lucene-8.11.2开发
* 部分库基本可用，单元测试补齐中

### 我们的目标

* 初期尽可能兼容Java版本Lucene的API接口
* 维护一套高质量的Go版本的搜索引擎库
* 提供比Java版本Lucene更强的性能

### 当前任务

* 完善基础库的单元测试
* 完善开发文档、设计文档
* 增加代码用例

### 项目概览

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

## 技术文档

### FST

* [1. 图解FST构造算法](https://juejin.cn/post/7311603506222088207)
* [2. FST构造-工程优化](https://juejin.cn/post/7311957969423663119)

## 尝试

> go1.21+

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
