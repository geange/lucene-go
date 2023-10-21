package simpletext

import "github.com/geange/lucene-go/core/index"

var _ index.DocValuesFormat = &DocValuesFormat{}

// DocValuesFormat
/**
plain text doc values format.
FOR RECREATIONAL USE ONLY
the .dat file contains the data. for numbers this is a "fixed-width" file, for example a single byte range:
   field myField
     type NUMERIC
     minvalue 0
     pattern 000
   005
   T
   234
   T
   123
   T
   ...

so a document's value (delta encoded from minvalue) can be retrieved by seeking to
startOffset + (1+pattern.length()+2)*docid. The extra 1 is the newline. The extra 2 is another newline
and 'T' or 'F': true if the value is real, false if missing. for bytes this is also a "fixed-width" file,
for example:

   field myField
     type BINARY
     maxlength 6
     pattern 0
   length 6
   foobar[space][space]
   T
   length 3
   baz[space][space][space][space][space]
   T
   ...

so a doc's value can be retrieved by seeking to startOffset + (9+pattern.length+maxlength+2)*doc the
extra 9 is 2 newlines, plus "length " itself. the extra 2 is another newline and 'T' or 'F': true
if the value is real, false if missing. for sorted bytes this is a fixed-width file, for example:

   field myField
     type SORTED
     numvalues 10
     maxLength 8
     pattern 0
     ordpattern 00
   length 6
   foobar[space][space]
   length 3
   baz[space][space][space][space][space]
   ...
   03
   06
   01
   10
   ...

so the "ord section" begins at startOffset + (9+pattern.length+maxlength)*numValues.
a document's ord can be retrieved by seeking to "ord section" + (1+ordpattern.length())*docid
an ord's value can be retrieved by seeking to startOffset + (9+pattern.length+maxlength)*ord
for sorted set this is a fixed-width file very similar to the SORTED case, for example:

   field myField
     type SORTED_SET
     numvalues 10
     maxLength 8
     pattern 0
     ordpattern XXXXX
   length 6
   foobar[space][space]
   length 3
   baz[space][space][space][space][space]
   ...
   0,3,5
   1,2

   10
   ...

 so the "ord section" begins at startOffset + (9+pattern.length+maxlength)*numValues.
 a document's ord list can be retrieved by seeking to "ord section" + (1+ordpattern.length())*docid
 this is a comma-separated list, and it's padded with spaces to be fixed width. so trim() and split() it.
 and beware the empty string! an ord's value can be retrieved by seeking to
 startOffset + (9+pattern.length+maxlength)*ord for sorted numerics, it's encoded (not very creatively)
 as a comma-separated list of strings the same as binary. beware the empty string! the reader can just
 scan this file when it opens, skipping over the data blocks and saving the offset/etc for each field.

 lucene.experimental
*/
type DocValuesFormat struct {
	name string
}

func NewSimpleTextDocValuesFormat() *DocValuesFormat {
	return &DocValuesFormat{name: "SimpleText"}
}

func (s *DocValuesFormat) GetName() string {
	return s.name
}

func (s *DocValuesFormat) FieldsConsumer(state *index.SegmentWriteState) (index.DocValuesConsumer, error) {
	return NewDocValuesWriter(state, "dat")
}

func (s *DocValuesFormat) FieldsProducer(state *index.SegmentReadState) (index.DocValuesProducer, error) {
	return NewDocValuesReader(state, "dat")
}
