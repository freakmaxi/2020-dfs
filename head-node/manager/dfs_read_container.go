package manager

import (
	"io"

	"github.com/freakmaxi/kertish-dfs/basics/common"
)

type ReadType int

const (
	RT_Folder ReadType = 1
	RT_File   ReadType = 2
)

type ReadContainer interface {
	Type() ReadType

	Folder() *common.Folder
	File() *common.File

	Read(w io.Writer, begins int64, ends int64) error
}

type readContainer struct {
	folder *common.Folder

	file          *common.File
	streamHandler func(w io.Writer, begins int64, ends int64) error
}

func newReadContainerForFolder(folder *common.Folder) ReadContainer {
	return &readContainer{
		folder: folder,
	}
}

func newReadContainerForFile(file *common.File, streamHandler func(w io.Writer, begins int64, ends int64) error) ReadContainer {
	return &readContainer{
		file:          file,
		streamHandler: streamHandler,
	}
}

func (r *readContainer) Type() ReadType {
	if r.file != nil {
		return RT_File
	}
	return RT_Folder
}

func (r *readContainer) Folder() *common.Folder {
	return r.folder
}

func (r *readContainer) File() *common.File {
	return r.file
}

func (r *readContainer) Read(w io.Writer, begins int64, ends int64) error {
	return r.streamHandler(w, begins, ends)
}

var _ ReadContainer = &readContainer{}
