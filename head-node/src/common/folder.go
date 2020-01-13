package common

import (
	"os"
	"sort"
	"strings"
	"time"
)

type Folder struct {
	Full     string        `json:"full"`
	Name     string        `json:"name"`
	Created  time.Time     `json:"created"`
	Modified time.Time     `json:"modified"`
	Folders  FolderShadows `json:"folders"`
	Files    Files         `json:"files"`
}

func NewFolder(folderPath string) *Folder {
	folderPath = CorrectPath(folderPath)
	_, name := Split(folderPath)

	return &Folder{
		Full:     folderPath,
		Name:     name,
		Created:  time.Now().UTC(),
		Modified: time.Now().UTC(),
		Folders:  make(FolderShadows, 0),
		Files:    make(Files, 0),
	}
}

func (f *Folder) NewFolder(name string, newFolderHandler func(*FolderShadow) error) error {
	name = CorrectPath(name)
	name = name[1:]

	if len(name) == 0 {
		return os.ErrInvalid
	}

	if strings.Index(name, pathSeparator) > -1 {
		return os.ErrInvalid
	}

	if f.exists(name) {
		return os.ErrExist
	}

	folderShadow := NewFolderShadow(Join(f.Full, name))
	if err := newFolderHandler(folderShadow); err != nil {
		return err
	}
	f.Folders = append(f.Folders, folderShadow)
	sort.Sort(f.Folders)

	return nil
}

func (f *Folder) NewFile(name string) (*File, error) {
	name = CorrectPath(name)
	name = name[1:]

	if len(name) == 0 {
		return nil, os.ErrInvalid
	}

	if strings.Index(name, pathSeparator) > -1 {
		return nil, os.ErrInvalid
	}

	if f.exists(name) {
		return nil, os.ErrExist
	}

	nf := newFile(name)
	f.Files = append(f.Files, nf)
	sort.Sort(f.Files)

	return nf, nil
}

func (f *Folder) Folder(name string) *string {
	for _, fs := range f.Folders {
		if strings.Compare(fs.Name, name) == 0 {
			folderFull := Join(f.Full, name)
			return &folderFull
		}
	}
	return nil
}

func (f *Folder) File(name string) *File {
	for _, sf := range f.Files {
		if strings.Compare(sf.Name, name) == 0 {
			return sf
		}
	}
	return nil
}

func (f *Folder) ReplaceFile(name string, file *File) {
	for i, sf := range f.Files {
		if strings.Compare(sf.Name, name) == 0 {
			if file == nil {
				f.Files = append(f.Files[:i], f.Files[i+1:]...)
			} else {
				f.Files[i] = file
			}
			sort.Sort(f.Files)
			return
		}
	}

	if file == nil {
		return
	}

	f.Files = append(f.Files, file)
	sort.Sort(f.Files)
}

func (f *Folder) DeleteFolder(name string, deleteFolderHandler func(string) error) error {
	for i, fs := range f.Folders {
		if strings.Compare(fs.Name, name) == 0 {
			if err := deleteFolderHandler(Join(f.Full, name)); err != nil {
				return err
			}
			f.Folders = append(f.Folders[:i], f.Folders[i+1:]...)
			sort.Sort(f.Folders)
			return nil
		}
	}
	return os.ErrNotExist
}

func (f *Folder) DeleteFile(name string, deleteFileHandler func(*File) error) error {
	for i, sf := range f.Files {
		if strings.Compare(sf.Name, name) == 0 {
			if err := deleteFileHandler(sf); err != nil {
				return err
			}
			f.Files = append(f.Files[:i], f.Files[i+1:]...)
			sort.Sort(f.Files)
			return nil
		}
	}
	return os.ErrNotExist
}

func (f *Folder) Size(sizeHandler func(FolderShadows) uint64) uint64 {
	s := uint64(0)

	for _, file := range f.Files {
		s += file.Size
	}

	if len(f.Folders) > 0 {
		s += sizeHandler(f.Folders)
	}

	return s
}

func (f *Folder) CloneInto(target *Folder) {
	if target == nil {
		return
	}

	copy(target.Folders, f.Folders)

	target.Files = make(Files, 0)
	for _, f := range f.Files {
		shadow := *f
		target.Files = append(target.Files, &shadow)
	}
}

func (f *Folder) Locked() bool {
	for _, file := range f.Files {
		if file.Locked {
			return true
		}
	}
	return false
}

func (f *Folder) exists(name string) bool {
	file := f.File(name)
	if file != nil {
		return true
	}

	folder := f.Folder(name)
	if folder != nil {
		return true
	}

	return false
}
