package main

import (
	"bufio"
	"fssync_db/internal/db"
	"fssync_db/internal/dir"
	"fssync_db/internal/file"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// user空間配下を再起的に取得
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// SQLファイルを事前に作成
	filesFile, err := os.Create("build/filesTable.sql")
	if err != nil {
		return
	}
	defer filesFile.Close()
	filesWriter := bufio.NewWriter(filesFile)
	defer filesWriter.Flush()

	dirsFile, err := os.Create("build/dirsTable.sql")
	if err != nil {
		return
	}
	defer dirsFile.Close()
	dirsWriter := bufio.NewWriter(dirsFile)
	defer dirsWriter.Flush()

	// バルクINSERTのヘッダー部分を先に書き込み
	filesWriter.WriteString(db.GetFileInsertHeader())
	dirsWriter.WriteString(db.GetDirInsertHeader())

	// 最初のエントリかどうかを追跡するフラグ
	isFirstFile := true
	isFirstDir := true

	processedCount := 0
	errorCount := 0

	filepath.WalkDir(home, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			errorCount++
			// エラーの詳細を出力（最初の10個まで）
			if errorCount <= 10 {
				println("エラー:", path, "->", err.Error())
			}
			// 権限エラーなどは無視して続行
			return nil
		}

		fileInfo, infoErr := info.Info()
		if infoErr != nil {
			return infoErr
		}

		// owner/group情報を取得
		// Windows以外でのみ動作する
		var userName string
		var groupName string
		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			uid := stat.Uid
			gid := stat.Gid
			// ユーザー名を取得
			u, userErr := user.LookupId(strconv.Itoa(int(uid)))
			userName = strconv.Itoa(int(uid)) // デフォルトはUID
			if userErr == nil {
				userName = u.Username
			}
			// グループ名を取得
			g, groupErr := user.LookupGroupId(strconv.Itoa(int(gid)))
			groupName = strconv.Itoa(int(gid)) // デフォルトはGID
			if groupErr == nil {
				groupName = g.Name
			}
		}

		parentPath := filepath.Dir(path)

		// dならDir構造体作成 & バルクINSERTの値部分を書き込み
		// それ以外ならFile構造体作成 & バルクINSERTの値部分を書き込み
		if info.IsDir() {
			newDir := dir.Dir{
				Path:    path,
				Perm:    fileInfo.Mode(),
				Parent:  parentPath,
				Owner:   userName,
				Group:   groupName,
				Size:    fileInfo.Size(),
				Updated: fileInfo.ModTime().Format(time.RFC3339),
				Name:    fileInfo.Name(),
			}
			// バルクINSERTの値部分を書き込み
			if isFirstDir {
				dirsWriter.WriteString(db.DirToFirstBulkValue(newDir))
				isFirstDir = false
			} else {
				dirsWriter.WriteString(db.DirToBulkValue(newDir))
			}
		} else {
			newFile := file.File{
				Perm:    fileInfo.Mode(),
				Parent:  parentPath,
				Owner:   userName,
				Group:   groupName,
				Size:    fileInfo.Size(),
				Updated: fileInfo.ModTime().Format(time.RFC3339),
				Name:    fileInfo.Name(),
			}
			// バルクINSERTの値部分を書き込み
			if isFirstFile {
				filesWriter.WriteString(db.FileToFirstBulkValue(newFile))
				isFirstFile = false
			} else {
				filesWriter.WriteString(db.FileToBulkValue(newFile))
			}
		}
		processedCount++
		return nil
	})

	// 処理結果を出力
	println("処理されたファイル・ディレクトリ数:", processedCount)
	println("エラー数:", errorCount)

	// バルクINSERT文を完結させるためのセミコロンを追加
	filesWriter.WriteString(";\n")
	dirsWriter.WriteString(";\n")

	// 明示的にFlushしてバッファの内容をファイルに書き込む
	filesWriter.Flush()
	dirsWriter.Flush()
}
