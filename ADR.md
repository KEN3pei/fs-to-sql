### 設計

Localのディレクトリを読み取ってテーブル構造に変換してDDLを生成する
- u: usersも生成する

### 命名規則など

Go Style
- https://google.github.io/styleguide/go/
Standard Go Project Layout
- https://github.com/golang-standards/project-layout/blob/master/README_ja.md

### memo
- package名,module名が長い場合それは切り分けられる可能性がある
- pathを環境変数として受け取って構造体に詰め替えるだけのpackageとそれを使ってDDLを生成するpackageに分けてもいい
- filetypeって色々あるがそれも全部読み込むのか？（一旦読み込むでいい）
  - 全てのfiletypeに対応しているかのtestは書きたいかも
- どうやって再起的に読み取るか
  - そのままトラバーサルするとO(N)で計算量と必要なメモリ量が増えてしまう？
  - packages
    - https://github.com/spf13/afero
    - https://github.com/d6o/GoTree
    - filepath.WalkDirやos.ReadDirで自作

### 参照記事

- bytes.Buffer型, strings.Reader型: https://zenn.dev/ken3pei/articles/8a68c730380432
  - bufからの読み取りとはOSのキャッシュのこと？bufとは？

- WalkDirFunc と path/filepath.WalkFunc の違い
  - https://pkg.go.dev/io/fs#WalkDirFunc
    - 第二引数の型が FileInfo ではなく DirEntry
    - この関数はディレクトリを読み取る前に呼び出され、SkipDir または SkipAll によってディレクトリの読み取りを完全にスキップするか、残りのファイルとディレクトリをすべてスキップすることを可能にします。
    - ディレクトリの読み取りに失敗した場合、そのディレクトリに対してエラーを報告するためにこの関数が再度呼び出されます。

- bitの左シフトなど計算方法について
  - 右記を参考にしたい: https://pkg.go.dev/io/fs#FileMode
  - マスク処理について: 
    - ModeType = ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice | ModeCharDevice | ModeIrregular

```shell
ModeDir        = 10000000 00000000 00000000 00000000
ModeSymlink    = 00001000 00000000 00000000 00000000
ModeNamedPipe  = 00000010 00000000 00000000 00000000
ModeSocket     = 00000001 00000000 00000000 00000000
ModeDevice     = 00000100 00000000 00000000 00000000
ModeCharDevice = 00000000 00100000 00000000 00000000
ModeIrregular  = 00000000 00001000 00000000 00000000
---------------------------------------------
ModeType (OR)  = 10001111 00101000 00000000 00000000
```

- os.Stat()とos.Lstat()の違い
  - https://zenn.dev/naoki_kuroda/articles/9d75c717a4d84a
  - os.Stat(): シンボリックリンク先のファイルなどの情報を返す
  - os.Lstat(): シンボリックリンク自体の情報を返す

### workDirの挙動

深さ優先探索で動作する。->これって効率的なんだっけ？（2pointer的な動かし方できない？）
つまりa~zの順でファイルとディレクトリを探しに行ってディレクトリなら階層を潜ってを繰り返す。

```go
func walkDir(path string, d fs.DirEntry, walkDirFn fs.WalkDirFunc) error {
    // 1. まず現在のパス（ディレクトリまたはファイル）を処理
    if err := walkDirFn(path, d, nil); err != nil || !d.IsDir() {
        return err
    }

    // 2. ディレクトリの場合、中身を読み取り
    // os.ReadDirはソートされて返す
    dirs, err := os.ReadDir(path)
    
    // 3. 各エントリを順番に再帰処理
    for _, d1 := range dirs {
        path1 := Join(path, d1.Name())
        if err := walkDir(path1, d1, walkDirFn); err != nil {
            // ...
        }
    }
}
```

### テーブルスキーマへの変換をどうやるか？

- sqlcのgenerateコマンドを活用する

配列を(1006,'田中','男’,90)といった形に変換する処理が必要

```sql
INSERT INTO directories (id,name,gender,point)
VALUES
(1006,'田中','男',90),
(1007,'土屋','女',55)
;
```

### 独自性

- localのfilesystemを監視して同期可能にする

### Goならわかるシステムプログラミング第2版

10.1 ファイルの変更監視(syscall.Inotify*)
- fsnotify package
  - Linux, Mac, Windowなどにマルチに対応している
  - Localのファイル変更はこれで受け取れる
- inotify_init(2)
  - 新規のinotifyインスタンスを初期化し、作成されたinotifyイベントキューに対応するファイルディスクリプタを返す。
- inotify_add_watch
  - 初期化済み inotify インスタンスに監視対象を追加する

MacOSにおけるwatchできる数の上限は？
$ ulimit -n
1048575

MacOSでは、FSEventsという仕組みによって1ディレクトリ=1fdなので効率が良い
・ファイルごとに個別にwatchする必要がない（→FD節約）
・逆に、一部のサブディレクトリだけを除外するのが難しい（設計上の制約）
しかし、Goのfsnotify packageはkqueueという仕組みを利用しており効率があまり良くない
MacOS専用のものなどもある。
・https://pkg.go.dev/github.com/fsnotify/fsevents
・こちらはFSEventsを利用する。
Linuxでも再起的にサブディレクトリの監視まで自動化してくれるpackage
・https://github.com/rjeczalik/notify
・LinuxではあくまでAddなどの処理をラップしているだけで仕組み自体は変わらない。
・MacOSの時はFSEventsを利用するので効率が良い

ファイルディスクリプタとは？
・ファイルに対するデータの通り道を識別するための目印
・「どのファイルにつながっているよ！」を示す目印
・以下のファイルディスクリプタ0~2は用途が最初から決まっている
0：標準入力
1：標準出力
2：標準エラー出力
・ファイルディスクリプタ (FD) は、「何かをオープンした瞬間にカーネルが割り振る」もので、普段は存在しません。
つまり「作らないと（open しないと）存在しない」。
　・ファイルアクセス時の流れ
   1. open("file.txt") → FD 3 が割り当てられる
   2. read(3, ...)
   3. close(3)

### FileModeの処理

dirならFileMode|0755の場合
=10000000000000000000000111101101 となり、1<<uint(32-1-0)で比較される。
=10000000000000000000000000000000 これを&でマスクすると1が返ってくる。これによってdという文字が対応づけられる。


### mapping

filepath.WorkDirなどでどれだけ情報を取得できるか？

path: directoryのみ自身を含めて記載
perm: DirEntry.Info()->FileInfo.Mode().String()[1:]
parent: *Dir
owner: fileInfo.Sys().(*syscall.Stat_t).Uid
group: fileInfo.Sys().(*syscall.Stat_t).Gid
size: FileInfo.Size()
updated: FileInfo.Modtime().Format() or .Unix(): string|int64
name: FileInfo.Name()

### time.Timeについて

Goでの時刻の扱い方
- https://zenn.dev/hsaki/articles/go-time-cheatsheet
- Unix() or Format()で変換して渡すのが良さそう

### GoでUMLを使う際のルール

https://developer.mamezou-tech.com/blogs/2024/07/01/uml-x-mapping-go1/
