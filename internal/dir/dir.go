package dir

import (
	"fmt"
	"os"
	"strconv"
)

type Dir struct {
	Path    string
	Perm    os.FileMode
	Parent  string
	Owner   string
	Group   string
	Size    int64
	Updated string
	Name    string
}

// どうやってtebleスキーマ生成する？
func (d *Dir) ToColumn() []string {
	// dir構造体をDDLに変換したい
	// path,perm,parent_path,owner,group,size,updated_at,name
	return []string{
		fmt.Sprintf("'%s'", d.Path),
		fmt.Sprintf("'%s'", d.Perm.String()),
		fmt.Sprintf("'%s'", d.Parent),
		fmt.Sprintf("'%s'", d.Owner),
		fmt.Sprintf("'%s'", d.Group),
		fmt.Sprintf("'%s'", strconv.FormatInt(d.Size, 10)),
		fmt.Sprintf("'%s'", d.Updated),
		fmt.Sprintf("'%s'", d.Name),
	}
}
