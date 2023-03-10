# Spotify file-based Client

[(ENGLISTH 英語)](./README.md)

## 特徴

**(1)** Spotify のプレイリスト情報をローカルにダウンロード.

- 楽曲 は テキストファイル
- プレイリスト は フォルダ  
  と表現されます.

<div style="text-align:center;">
↓
</div>

**(2)** バックアップ, 曲の移動, 曲の追加などを行う.  
 (エクスプローラーなどで操作できます)

<div style="text-align:center;">
↓
</div>

**(3)** フォルダ,テキストファイルの状態を Spotify に転送

<div style="text-align:center;">
楽曲のテキストファイルの例
</div>

```txt
id 6ODZT1FGTE2q4two05giS1
name 新宝島
artist sakanaction
album 魚図鑑
seconds 302493
isrc JPVI01527970
file_name 新宝島.txt
```

## How to use

### (1) Download

https://github.com/kajikentaro/spotify-fbc/releases/tag/0.1  
ここから Windows 版と Linux 版がダウンロードできます.

ダウンロードが完了したら, `spotify-fbc`に名前変更することをおすすめします.(Windows の場合は `spotify-fbc.exe`)

`spotify-fbc`が存在するファイル内で以下コマンドを実行し,help が表示されることを確認しましょう

```
$ spotify-fbc help
Spotify file-based client:
Edit your playlists by moving directories and file locations

Usage:
  spotify-fbc [command]

Available Commands:
  clean       Clean up unused playlist entity txt
  compare     Compare local playlists with your spotify account and print the difference
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  login       Perform login process
  logout      Logout from your spotify account excluding API keys
  pull        Download playlists that your spotify account has. All of your existing local playlists will be overwritten
  push        Synchronize your local files and directories with your spotify account
  reset       Delete user-specific data such as OAuth token and Client ID excluding music txt
  version     Print the version number of spotify-fbc

Flags:
  -h, --help   help for spotify-fbc
```

### (2) 下準備(API キーの作成)

1. https://developer.spotify.com/dashboard/applications  
   にアクセスし, `CREATE AN APP`ボタンをクリックします.
2. `App name`, `App description`を適当に決めます.
3. `Client ID` をメモします.
4. `SHOW CLIENT SECRET`ボタンを押し, 表示される`Client Secret`をメモします.
5. `EDIT SETTINGS`ボタンを押し,`Redirect URIs`に`http://localhost:8080/callback`を入力し,`ADD`をクリックします.  
   その後,`SAVE`で保存します.

### (3) ログイン

以下コマンドを実行します

```
$ spotify-fbc-0.1-linux-amd64 login
```

1. `Enter your Client ID:`と表示されたら`Client ID`を入力します
2. `Enter your Cilent Secret:`と表示されたら`Client Secret`を入力します
3. `Enter your Redirect URI:`と表示されたらそのままエンターキーを押します.(必要なら変更したものを入力してください)
4. URL が表示されるのでブラウザで開き, Spotify のログインを行います
5. `token cache was saved to ~~~`が表示されたら成功です

### (4) Spotify 楽曲情報のダウンロード

以下コマンドを実行します

```
$ spotify-fbc pull
```

spotify-fbc というディレクトリが生成され, その中に楽曲,プレイリスト情報が保存されます

### (5) 差分情報の比較

必要な編集を`spotify-fbc`ディレクトリに行ったら, `push`を行う前に差分を確認しましょう

```
$ spotify-fbc compare

+ new playlist
  + 新宝島
- deleted playlist
```

`new playlist`プレイリストと, `新宝島`という楽曲が追加され,  
`deleted playlist`が削除されるということが確認できます.

### (6) Spotify 楽曲情報のアップロード

以下コマンドを実行します

```
$ spotify-fbc push
```

既存の Spotify プレイリストが, `spotify-fbc`ディレクトリに完全に置き換わります.
実行は慎重に!

### (7) 楽曲検索

楽曲情報テキストファイルは, いくつかのプロパティから構成されています

例えば以下のようなファイルを新規作成し,`spotify-fbc push`を行うと,

```text
name 新宝島
artist sakanaction
```

楽曲名を`新宝島`, アーティスト名を`sakanaction`として検索が行われます.

その他,`id`, `name`, `artist`,`album`, `isrc`プロパティが検索に対応しています.
(他のプロパティはシステムの管理用です)
