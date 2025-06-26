# Minecraft Forge Server Template

このディレクトリは、Minecraft Forge サーバー用のテンプレートです。

## ディレクトリ構造

```
forge/
├── Dockerfile              # Forge サーバー用のDockerfile
├── server.properties       # サーバー設定ファイル
├── ops.json                # オペレーター（管理者）リスト
├── whitelist.json          # ホワイトリスト
├── mods/                   # MODファイルを配置するディレクトリ
├── config/                 # MOD設定ファイルが保存されるディレクトリ
├── world/                  # ワールドデータが保存されるディレクトリ
└── README.md               # このファイル
```

## 使用方法

### 1. 設定ファイルの編集

- `server.properties`: サーバーの基本設定（ポート、MOTD、ゲームモードなど）
- `ops.json`: 管理者権限を持つプレイヤーのリスト
- `whitelist.json`: サーバーにアクセス可能なプレイヤーのリスト

### 2. MODの追加

1. `mods/` ディレクトリに `.jar` 形式のMODファイルを配置
2. 必要に応じて `config/` ディレクトリ内の設定ファイルを編集

### 3. Dockerfileの設定

`Dockerfile` 内の以下の項目を必要に応じて変更:

- `VERSION`: Minecraftのバージョン
- `MINECRAFT_VERSION`: Minecraftのバージョン（VERSIONと同じ）
- `FORGE_VERSION`: Forgeのバージョン
- `MEMORY`: サーバーに割り当てるメモリ

### 4. サーバーの起動

テンプレートディレクトリ（`minecraft/template/`）から:

```bash
# Forgeサーバーのみ起動
docker-compose up forge-server

# すべてのサーバーを起動
docker-compose up

# バックグラウンドで起動
docker-compose up -d forge-server
```

## 注意事項

- 初回起動時は、Forgeサーバーのダウンロードとインストールに時間がかかります
- `ops.json` と `whitelist.json` の例では、ダミーのUUIDとプレイヤー名を使用しています。実際のプレイヤー情報に置き換えてください
- MODによっては、サーバー側とクライアント側の両方にインストールが必要な場合があります
- `world/` ディレクトリは初回起動時に自動的に生成されます

## トラブルシューティング

- サーバーが起動しない場合は、`docker-compose logs forge-server` でログを確認してください
- MODの競合がある場合は、`config/` ディレクトリ内の設定ファイルを確認し、必要に応じて調整してください
