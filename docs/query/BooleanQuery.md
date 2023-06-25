# BooleanQuery

## TermQuery

the index data like that 

```json
{
  "0": {
    "content": ["h"],
    "author": "author1"
  },
  "1": {
    "content": ["b"],
    "author": "author2"
  },
  "2": {
    "content": ["a", "c"],
    "author": "author3"
  },
  "3": {
    "content": ["a", "c", "e"],
    "author": "author4"
  },
  "4": {
    "content": ["h"],
    "author": "author5"
  },
  "5": {
    "content": ["c", "e"],
    "author": "author6"
  },
  "6": {
    "content": ["c", "a", "e"],
    "author": "author7"
  },
  "7": {
    "content": ["f"],
    "author": "author8"
  },
  "8": {
    "content": ["b", "c", "d", "e", "c", "e"],
    "author": "author9"
  },
  "9": {
    "content": ["a", "c", "e", "a", "b", "c"],
    "author": "author10"
  }
}
```

```
q1 := search.NewTermQuery(index.NewTerm("content", []byte("a")))
q2 := search.NewTermQuery(index.NewTerm("content", []byte("c")))
q3 := search.NewTermQuery(index.NewTerm("content", []byte("e")))

builder := search.NewBooleanQueryBuilder()
builder.AddQuery(q1, search.OccurMust)
builder.AddQuery(q2, search.OccurMust)
builder.AddQuery(q3, search.OccurMust)
query, err := builder.Build()
if err != nil {
    panic(err)
}

topDocs, err := searcher.SearchTopN(query, 5)
if err != nil {
    panic(err)
}

for i, doc := range topDocs.GetScoreDocs() {
    fmt.Printf("result%d: 文档%d\n", i, doc.GetDoc())
}
```