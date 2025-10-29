package db

import (
	"fmt"
	"fssync_db/internal/dir"
	"fssync_db/internal/file"
	"strings"
)

// 単一のディレクトリエントリ用のINSERT文を生成
func DirToSingleInsertQuery(d dir.Dir) string {
	return fmt.Sprintf("INSERT INTO directories (\"path\",\"perm\",\"parent_path\",\"owner\",\"group\",\"size\",\"updated_at\",\"name\") VALUES (%s);\n",
		strings.Join(d.ToColumn(), ","))
}

// 単一のファイルエントリ用のINSERT文を生成
func FileToSingleInsertQuery(f file.File) string {
	return fmt.Sprintf("INSERT INTO files (\"path\",\"perm\",\"parent_path\",\"owner\",\"group\",\"size\",\"updated_at\",\"name\") VALUES (%s);\n",
		strings.Join(f.ToColumn(), ","))
}

// バルクINSERT用のヘッダー部分を生成
func GetDirInsertHeader() string {
	return "INSERT INTO directories (\"path\",\"perm\",\"parent_path\",\"owner\",\"group\",\"size\",\"updated_at\",\"name\") VALUES\n"
}

func GetFileInsertHeader() string {
	return "INSERT INTO files (\"path\",\"perm\",\"parent_path\",\"owner\",\"group\",\"size\",\"updated_at\",\"name\") VALUES\n"
}

// バルクINSERT用の値部分を生成（最初のエントリ）
func DirToFirstBulkValue(d dir.Dir) string {
	return fmt.Sprintf("(%s)", strings.Join(d.ToColumn(), ","))
}

func FileToFirstBulkValue(f file.File) string {
	return fmt.Sprintf("(%s)", strings.Join(f.ToColumn(), ","))
}

// バルクINSERT用の値部分を生成（2番目以降のエントリ）
func DirToBulkValue(d dir.Dir) string {
	return fmt.Sprintf(",\n(%s)", strings.Join(d.ToColumn(), ","))
}

func FileToBulkValue(f file.File) string {
	return fmt.Sprintf(",\n(%s)", strings.Join(f.ToColumn(), ","))
}
