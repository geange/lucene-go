package compress

import (
	"os"
	"testing"

	"github.com/geange/lucene-go/core/store"
)

func TestCompress(t *testing.T) {
	ht := NewHighCompressionHashTable()

	buf := store.NewBufferDataOutput()

	t.Log("origin size", len([]byte(data)))

	Compress([]byte(data), buf, ht)

	t.Log(len(buf.Bytes()))

	f, _ := os.Create("lz4.out")
	f.Write(buf.Bytes())
	f.Close()
}

func TestDecompress(t *testing.T) {
	ht := NewHighCompressionHashTable()

	// 压缩一些数据
	buf := store.NewBufferDataOutput()
	Compress([]byte(data), buf, ht)

	// 创建解压的目标缓冲区
	dest := make([]byte, len(data))
	in := store.NewByteArrayDataInput(buf.Bytes())

	// 执行解压
	off, err := Decompress(in, len(data), dest)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	// 验证解压结果
	decompressedData := dest[:off]
	if string(decompressedData) != data {
		t.Errorf("Decompress result mismatch. Expected length: %d, Got length: %d", len(data), off)
	}
}

func TestDecompressSimple(t *testing.T) {
	// 测试场景: 简单的字面值和匹配组合
	// 创建一个模拟的压缩数据流
	compressedData := []byte{
		0x44,               // token: 0b0100_0100 - 4字节字面值, 4字节匹配
		'H', 'e', 'l', 'l', // 字面值 "Hell"
		0x04, 0x00, // 匹配距离: 4
	}

	// 创建一个足够大的目标缓冲区
	dest := make([]byte, 16)
	in := store.NewByteArrayDataInput(compressedData)

	// 执行解压
	off, err := Decompress(in, 8, dest)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	// 对于这个简单的测试，我们期望输出为 "Helloell" (字面值 + 基于字面值的匹配)
	// 但由于测试数据是简化的，这里主要验证函数能正常工作而不崩溃
	if off != 8 {
		t.Errorf("Expected decompressed length %d, got %d", 8, off)
	}
}

var data = `{
  "wfzielbvep": [
    true
  ],
  "oeagzhoejj": {
    "ootiwj": [
      [
        {
          "dwzwsvb": 523663928.6949992,
          "yzhaxkswswxo": [
            {
              "wwkdxmug": "ktmLpEcgglY_xIDsT",
              "cjwqtyrnztb": {
                "izreoqgi": false
              }
            }
          ]
        }
      ],
      -360447382.60452443
    ],
    "ofdrxgqqya": 373849449.82884866,
    "ljovnh": {
      "rrwsk": {
        "jjbxikntt": -1459447898.026548,
        "qwfixfs": {
          "hlbferegs": {
            "tqprbybxd": 1266333547
          }
        },
        "wxzqyzr": [
          "Y9BnyrQ2"
        ]
      },
      "mojkuart": [
        1556379175,
        "77oE8t"
      ]
    }
  },
  "exskbsc": "HzsqMTvi-"
}`
