package packed

import (
	"errors"
	"github.com/geange/lucene-go/core/util/packed/bulkoperation"
	"github.com/geange/lucene-go/core/util/packed/common"
)

var (
	packedBulkOps = []common.BulkOperation{
		bulkoperation.NewPacked1(),
		bulkoperation.NewPacked2(),
		bulkoperation.NewPacked3(),
		bulkoperation.NewPacked4(),
		bulkoperation.NewPacked5(),
		bulkoperation.NewPacked6(),
		bulkoperation.NewPacked7(),
		bulkoperation.NewPacked8(),
		bulkoperation.NewPacked9(),
		bulkoperation.NewPacked10(),
		bulkoperation.NewPacked11(),
		bulkoperation.NewPacked12(),
		bulkoperation.NewPacked13(),
		bulkoperation.NewPacked14(),
		bulkoperation.NewPacked15(),
		bulkoperation.NewPacked16(),
		bulkoperation.NewPacked17(),
		bulkoperation.NewPacked18(),
		bulkoperation.NewPacked19(),
		bulkoperation.NewPacked20(),
		bulkoperation.NewPacked21(),
		bulkoperation.NewPacked22(),
		bulkoperation.NewPacked23(),
		bulkoperation.NewPacked24(),
		bulkoperation.NewPacked(25),
		bulkoperation.NewPacked(26),
		bulkoperation.NewPacked(27),
		bulkoperation.NewPacked(28),
		bulkoperation.NewPacked(29),
		bulkoperation.NewPacked(30),
		bulkoperation.NewPacked(31),
		bulkoperation.NewPacked(32),
		bulkoperation.NewPacked(33),
		bulkoperation.NewPacked(34),
		bulkoperation.NewPacked(35),
		bulkoperation.NewPacked(36),
		bulkoperation.NewPacked(37),
		bulkoperation.NewPacked(38),
		bulkoperation.NewPacked(39),
		bulkoperation.NewPacked(40),
		bulkoperation.NewPacked(41),
		bulkoperation.NewPacked(42),
		bulkoperation.NewPacked(43),
		bulkoperation.NewPacked(44),
		bulkoperation.NewPacked(45),
		bulkoperation.NewPacked(46),
		bulkoperation.NewPacked(47),
		bulkoperation.NewPacked(48),
		bulkoperation.NewPacked(49),
		bulkoperation.NewPacked(50),
		bulkoperation.NewPacked(51),
		bulkoperation.NewPacked(52),
		bulkoperation.NewPacked(53),
		bulkoperation.NewPacked(54),
		bulkoperation.NewPacked(55),
		bulkoperation.NewPacked(56),
		bulkoperation.NewPacked(57),
		bulkoperation.NewPacked(58),
		bulkoperation.NewPacked(59),
		bulkoperation.NewPacked(60),
		bulkoperation.NewPacked(61),
		bulkoperation.NewPacked(62),
		bulkoperation.NewPacked(63),
		bulkoperation.NewPacked(64),
	}

	packedSingleBlockBulkOps = []common.BulkOperation{
		bulkoperation.NewPackedSingleBlock(1),
		bulkoperation.NewPackedSingleBlock(2),
		bulkoperation.NewPackedSingleBlock(3),
		bulkoperation.NewPackedSingleBlock(4),
		bulkoperation.NewPackedSingleBlock(5),
		bulkoperation.NewPackedSingleBlock(6),
		bulkoperation.NewPackedSingleBlock(7),
		bulkoperation.NewPackedSingleBlock(8),
		bulkoperation.NewPackedSingleBlock(9),
		bulkoperation.NewPackedSingleBlock(10),
		nil,
		bulkoperation.NewPackedSingleBlock(12),
		nil,
		nil,
		nil,
		bulkoperation.NewPackedSingleBlock(16),
		nil,
		nil,
		nil,
		nil,
		bulkoperation.NewPackedSingleBlock(21),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		bulkoperation.NewPackedSingleBlock(32),
	}
)

func Of(format Format, bitsPerValue int) (common.BulkOperation, error) {
	switch format.(type) {
	case *formatPacked:
		if packedBulkOps[bitsPerValue-1] != nil {
			return packedBulkOps[bitsPerValue-1], nil
		}
	case *formatPackedSingleBlock:
		if packedSingleBlockBulkOps[bitsPerValue-1] != nil {
			return packedSingleBlockBulkOps[bitsPerValue-1], nil
		}
	}
	return nil, errors.New("AssertionError")
}
