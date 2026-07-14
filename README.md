# asc — Apple App Store Connect API CLI

App Store Connect APIを操作するCLIツール。AWS CLIと同様のプロファイル方式でクレデンシャルを管理します。

## インストール

```bash
go build -o asc .
# または
go install github.com/ideamans/apple-app-store-connect@latest
```

## セットアップ

App Store Connectでダウンロードした `.p8` キーとIssuer IDを登録します。

```bash
asc configure --issuer-id <ISSUER_ID> --key ~/Downloads/AuthKey_XXXXXXXXXX.p8
```

- キーは `~/.config/apple-app-store-connect/keys/` にコピーされ（パーミッション0600）、プロファイルが `config.toml` に登録されます
- Key IDはファイル名 `AuthKey_XXXXXXXXXX.p8` から自動導出されます（`--key-id` で明示指定も可能）
- 最初に登録したプロファイルがデフォルトになります
- 別チームのキーは `--profile <name>` を付けて登録します

```bash
asc configure --profile client-a --issuer-id <ISSUER_ID> --key ~/Downloads/AuthKey_YYYYYYYYYY.p8
```

## プロファイル管理

```bash
asc profiles list          # 一覧（デフォルトに印が付く）
asc profiles use client-a  # デフォルトを切り替え
asc profiles remove NAME   # プロファイル削除（キーファイルは残る）
```

## 使い方

```bash
asc apps list                    # アプリ一覧（認証確認を兼ねる）
asc --profile client-a apps list # プロファイル指定

# 未対応のエンドポイントは汎用コマンドで
asc api /v1/apps
asc api "/v1/apps?filter[bundleId]=com.example.app"
asc api -X POST /v1/betaGroups -d @payload.json

# curl等で使うJWTを発行
curl -H "Authorization: Bearer $(asc token)" https://api.appstoreconnect.apple.com/v1/apps
```

## ヘルプ

- `asc --help` / `asc <command> --help` — 人間向けの通常のヘルプ
- `asc --llm` — LLMエージェント向けの詳細リファレンスを一括出力（クレデンシャルモデル、解決順序、全コマンド・フラグ、`asc api` のJSON:API/ページネーションの注意点まで含む）。どのサブコマンドに付けても同じ全文が出ます

## クレデンシャルの解決順序

1. 環境変数 `ASC_ISSUER_ID` + `ASC_PRIVATE_KEY_PATH`（またはCI向けに `ASC_PRIVATE_KEY_BASE64` + `ASC_KEY_ID`）
2. `--profile` フラグ → 環境変数 `ASC_PROFILE` → `config.toml` の `default_profile`

`.env` ファイルは読み込みません。direnv等でシェル側で環境変数化してください。

## 設定ファイル

```
~/.config/apple-app-store-connect/
├── config.toml   # プロファイル定義（0600）
└── keys/         # p8秘密鍵（0700 / 各ファイル0600）
    └── AuthKey_XXXXXXXXXX.p8
```

```toml
default_profile = "default"

[profiles.default]
issuer_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
key_id = "XXXXXXXXXX"
private_key = "keys/AuthKey_XXXXXXXXXX.p8"
```
