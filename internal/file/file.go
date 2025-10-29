package file

import (
	"fmt"
	"os"
	"strconv"
)

type File struct {
	Perm    os.FileMode
	Parent  string
	Owner   string
	Group   string
	Size    int64
	Updated string
	Name    string
}

// どうやってtebleスキーマ生成する？
func (f *File) ToColumn() []string {
	// dir構造体をDDLに変換したい
	// path,perm,parent_path,owner,group,size,updated_at,name
	return []string{
		fmt.Sprintf("'%s/%s'", f.Parent, f.Name),
		fmt.Sprintf("'%s'", f.Perm.String()),
		fmt.Sprintf("'%s'", f.Parent),
		fmt.Sprintf("'%s'", f.Owner),
		fmt.Sprintf("'%s'", f.Group),
		fmt.Sprintf("'%s'", strconv.FormatInt(f.Size, 10)),
		fmt.Sprintf("'%s'", f.Updated),
		fmt.Sprintf("'%s'", f.Name),
	}
}
