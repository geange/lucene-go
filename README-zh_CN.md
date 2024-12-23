# lucene-go

[![GoDoc](https://godoc.org/github.com/geange/lucene-go?status.svg)](https://godoc.org/github.com/geange/lucene-go)
[![Go](https://github.com/geange/lucene-go/actions/workflows/go.yml/badge.svg)](https://github.com/geange/lucene-go/actions/workflows/go.yml)
[![CodeQL](https://github.com/geange/lucene-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/geange/lucene-go/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/geange/lucene-go/graph/badge.svg?token=52HZJSPPS6)](https://codecov.io/gh/geange/lucene-go)

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

#### IndexWriter

Using `IndexWriter`

```go
package main

import (
  "context"
  "fmt"
  "os"

  "github.com/geange/lucene-go/codecs/simpletext"
  "github.com/geange/lucene-go/core/document"
  "github.com/geange/lucene-go/core/index"
  "github.com/geange/lucene-go/core/search"
  "github.com/geange/lucene-go/core/store"
)

func main() {
  err := os.RemoveAll("data")
  if err != nil {
    panic(err)
  }

  dir, err := store.NewNIOFSDirectory("data")
  if err != nil {
    panic(err)
  }

  codec := simpletext.NewCodec()
  similarity, err := search.NewBM25Similarity()
  if err != nil {
    panic(err)
  }

  config := index.NewIndexWriterConfig(codec, similarity)

  ctx := context.Background()

  writer, err := index.NewIndexWriter(ctx, dir, config)
  if err != nil {
    panic(err)
  }
  defer writer.Close()

  {
    doc := document.NewDocument()
    doc.Add(document.NewTextField("a", "74", true))
    doc.Add(document.NewTextField("b", "86", true))
    doc.Add(document.NewTextField("c", "1237", true))
    docID, err := writer.AddDocument(ctx, doc)
    if err != nil {
      panic(err)
    }
    fmt.Println("add new document:", docID)
  }

  {
    doc := document.NewDocument()
    doc.Add(document.NewTextField("a", "74", true))
    doc.Add(document.NewTextField("b", "123", true))
    doc.Add(document.NewTextField("c", "789", true))

    docID, err := writer.AddDocument(context.Background(), doc)
    if err != nil {
      panic(err)
    }
    fmt.Println("add new document:", docID)
  }

  {
    doc := document.NewDocument()
    doc.Add(document.NewTextField("a", "741", true))
    doc.Add(document.NewTextField("b", "861", true))
    doc.Add(document.NewTextField("c", "12137", true))
    docID, err := writer.AddDocument(context.Background(), doc)
    if err != nil {
      panic(err)
    }
    fmt.Println("add new document:", docID)
  }
}

```

#### IndexReader

```go
package main

import (
	"context"
	"fmt"

	_ "github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/index"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/store"
)

func main() {
	dir, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}

	reader, err := index.OpenDirectoryReader(context.Background(), dir, nil, nil)
	if err != nil {
		panic(err)
	}

	searcher, err := search.NewIndexSearcher(reader)
	if err != nil {
		panic(err)
	}

	query := search.NewTermQuery(index.NewTerm("a", []byte("74")))
	builder := search.NewBooleanQueryBuilder()
	builder.AddQuery(query, index2.OccurMust)
	booleanQuery, err := builder.Build()
	if err != nil {
		panic(err)
	}

	topDocs, err := searcher.SearchTopN(context.Background(), booleanQuery, 2)
	if err != nil {
		panic(err)
	}

	result := topDocs.GetScoreDocs()
	for _, scoreDoc := range result {
		docID := scoreDoc.GetDoc()
		document, err := reader.Document(context.Background(), docID)
		if err != nil {
			panic(err)
		}

		values := make([]any, 0)
		for _, field := range document.Fields() {
			values = append(values, fmt.Sprintf("name:%s, value:%v", field.Name(), field.Get()))
		}
		fmt.Println("docId: ", scoreDoc.GetDoc(), "values", values)
	}
}

```