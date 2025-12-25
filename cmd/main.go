package main

import (
	"bufio"
	"fssync_db/internal/db"
	"fssync_db/internal/dir"
	"fssync_db/internal/file"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
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

	type Job struct {
		data db.Schema
	}

	// job worker起動
	jobs := make(chan Job, 10)

	// 最初のエントリかどうかを追跡するフラグ
	isFirstFile := true
	isFirstDir := true

	// ワーカーを先に起動
	processedCount := 0
	go func() {
		for job := range jobs { // jobs が閉じるまでループ
			if dir, ok := job.data.(dir.Dir); ok {
				// バルクINSERTの値部分を書き込み
				if isFirstDir {
					dirsWriter.WriteString(db.DirToFirstBulkValue(dir))
					isFirstDir = false
				} else {
					dirsWriter.WriteString(db.DirToBulkValue(dir))
				}
			}
			if file, ok := job.data.(file.File); ok {
				// バルクINSERTの値部分を書き込み
				if isFirstFile {
					filesWriter.WriteString(db.FileToFirstBulkValue(file))
					isFirstFile = false
				} else {
					filesWriter.WriteString(db.FileToBulkValue(file))
				}
			}
			processedCount++
		}
	}()

	errorCount := 0
	// syscallを減らすためのキャッシュを定義
	userCache := make(map[uint32]string)
	groupCache := make(map[uint32]string)

	// WaitGroupで全てのgoroutineの完了を待つ
	var wg sync.WaitGroup

	filepath.WalkDir(home, func(path string, info fs.DirEntry, err error) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err != nil {
				errorCount++
				// エラーの詳細を出力（最初の10個まで）
				if errorCount <= 10 {
					println("エラー:", path, "->", err.Error())
				}
				// 権限エラーなどは無視して続行
				return
			}

			fileInfo, infoErr := info.Info()
			if infoErr != nil {
				log.Fatal(infoErr.Error())
				return
			}

			// owner/group情報を取得
			// Windows以外でのみ動作する
			var userName string
			var groupName string
			if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
				uid := stat.Uid
				gid := stat.Gid
				// usernameを取得(キャッシュ使用)
				if cachedName, ok := userCache[uid]; ok {
					userName = cachedName
				} else {
					u, err := user.LookupId(strconv.Itoa(int(uid)))
					if err == nil {
						userName = u.Username
						userCache[uid] = userName
					}
				}
				// グループ名を取得(キャッシュ使用)
				if cachedGroupName, ok := groupCache[gid]; ok {
					groupName = cachedGroupName
				} else {
					g, err := user.LookupGroupId(strconv.Itoa(int(gid)))
					groupName = strconv.Itoa(int(gid)) // デフォルトはGID
					if err == nil {
						groupName = g.Name
					}
					groupCache[gid] = groupName // 成功・失敗に関わらずキャッシュ
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
				jobs <- Job{
					data: newDir,
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
				jobs <- Job{
					data: newFile,
				}
			}
		}()
		return nil
	})

	// 全てのgoroutineが完了するのを待つ
	wg.Wait()

	// jobs channelをclose(バッファに残った処理の完了後終了)
	close(jobs)

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
