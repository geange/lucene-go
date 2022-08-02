package store

// A Directory provides an abstraction layer for storing a list of files. A directory contains only files
// (no sub-folder hierarchy). Implementing classes must comply with the following:
//
// * A file in a directory can be created (createOutput), appended to, then closed.
// * A file open for writing may not be available for read access until the corresponding IndexOutput is closed.
// * Once a file is created it must only be opened for input (openInput), or deleted (deleteFile). Calling
// 	 createOutput on an existing file must throw java.nio.file.FileAlreadyExistsException.
//
// See Also: FSDirectory, RAMDirectory, FilterDirectory
type Directory interface {
}
