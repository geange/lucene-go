# lucene-go

[![GoDoc](https://godoc.org/github.com/geange/lucene-go?status.svg)](https://godoc.org/github.com/geange/lucene-go)
[![Go](https://github.com/geange/lucene-go/actions/workflows/go.yml/badge.svg)](https://github.com/geange/lucene-go/actions/workflows/go.yml)
[![CodeQL](https://github.com/geange/lucene-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/geange/lucene-go/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/geange/lucene-go/graph/badge.svg?token=52HZJSPPS6)](https://codecov.io/gh/geange/lucene-go)

**[中文](README-zh_CN.md)**

## About

Lucene is a search engine library. `lucene-go` is its Golang version implementation.

### Current Version

* Only support Go1.21+
* Developed based on Lucene 8.11.2
* Some libraries are basically available, and unit testing is being completed

### Our Goals

* API interface compatible with Java version Lucene
* Maintain a high-quality Go version search engine library
* Provides stronger performance than the Java version of Lucene

### Current Tasks

* Improve the unit testing of the basic library
* Improve development and design documents
* Add Code Use Cases

### Project Overview

The goal of the project has undergone several twists and turns during the development process, encountering far greater
difficulties than expected, language differences, and a lack of theoretical knowledge. After a year of development, the
following major modules have been gradually completed:

* core/store: Lucene's storage module is mainly responsible for data serialization processing
* core/document: Define some search related data structures for Lucene
* core/index: The implementation of Lucene index is also the main package exposed to the public
* core/search: Mainly including the implementation of query (used for data retrieval in indexes)
* memory: Implemented a memory based search engine, a simplified version of Lucene
* util/fst: Implementation of FST (Lucene's important data structure)
* util/automaton: Implementation of Automata
* codes: serialization, currently only supports the format of simpleText (using plain text information to record index
  data)

> It should be noted that the current project is not perfect! Do not use for any project~

## Try

> go1.21+

### Example

[More Example](https://github.com/geange/lucene-go-example)

Using `IndexWriter`

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

  codec := simpletext.NewCodec()
  similarity, err := search.NewCastBM25Similarity()

  config := index.NewIndexWriterConfig(codec, similarity)

  writer, err := index.NewIndexWriter(context.Background(), dir, config)
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
    doc.Add(document.NewStoredField[int32]("a", 74))
    doc.Add(document.NewStoredField[int32]("a1", 86))
    doc.Add(document.NewStoredField[int32]("a2", 1237))
    docID, err := writer.AddDocument(context.Background(), doc)
    if err != nil {
      panic(err)
    }
    fmt.Println(docID)
  }

  {
    doc := document.NewDocument()
    doc.Add(document.NewStoredField[int32]("a", 123))
    doc.Add(document.NewStoredField[int32]("a1", 123))
    doc.Add(document.NewStoredField[int32]("a2", 789))

    docID, err := writer.AddDocument(context.Background(), doc)
    if err != nil {
      panic(err)
    }
    fmt.Println(docID)
  }

  {
    doc := document.NewDocument()
    doc.Add(document.NewStoredField[int32]("a", 741))
    doc.Add(document.NewStoredField[int32]("a1", 861))
    doc.Add(document.NewStoredField[int32]("a2", 12137))
    docID, err := writer.AddDocument(context.Background(), doc)
    if err != nil {
      panic(err)
    }
    fmt.Println(docID)
  }
}

```
