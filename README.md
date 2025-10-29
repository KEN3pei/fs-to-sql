### 概要

自身の$HOME配下の全てのディレクトリ、ファイルの情報を使ってDBデータを生成するツール。

適当なデータを使ってクエリを検証したい場合などに、テーブル構成を考えたりするのが手間だったり、そこそこのデータ量へのクエリを検証したい場合に使用することを想定している。

### 使用方法

1. go run cmd/main.goを実行

- 権限エラーなどあれば表示される
- 処理したファイルとディレクトリの数がわかる

```shell
$ go run cmd/main.go
エラー: /Users/uenokensuke/Library/Application Support/FileProvider -> open /Users/uenokensuke/Library/Application Support/FileProvider: operation not permitted
エラー: /Users/uenokensuke/Library/Application Support/Knowledge -> open /Users/uenokensuke/Library/Application Support/Knowledge: operation not permitted
処理されたファイル・ディレクトリ数: 1454623
エラー数: 58
```

2. build配下に2つの.sqlファイルが生成されるのでこれを手元のDBサーバーにimportする
