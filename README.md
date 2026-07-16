# asc — Apple App Store Connect API CLI

App Store Connect APIを操作するCLIツール。AWS CLIと同様のプロファイル方式でクレデンシャルを管理します。

## インストール

```bash
go build ./cmd/asc          # カレントに asc を出力
# または
go install github.com/ideamans/apple-app-store-connect-cli/cmd/asc@latest
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

## 審査提出フロー用コマンド

メタデータ入力から審査提出までの主要操作を、`asc api` を内部で叩く薄いラッパとして提供します。`--app` はアプリID・Bundle IDのどちらでも指定できます。書き込み系はすべて `--dry-run` で実リクエストをプレビューできます（送信せず標準エラーへ出力）。

```bash
# App情報（名前・サブタイトル・プライバシーURL / カテゴリ / 年齢レーティング）
asc appinfo show --app <APP>
asc appinfo localize --app <APP> --locale ja --name "..." --subtitle "..." --privacy-url https://...
asc appinfo category --app <APP> --primary PRODUCTIVITY --secondary BUSINESS
asc appinfo age-rating --app <APP> --attrs @agerating.json

# バージョン（作成 / ローカライズ / ビルド選択）
asc version list --app <APP>
asc version create --app <APP> --version 1.0
asc version localize --app <APP> --locale ja --description @desc.txt --keywords "領収書,レシート,Excel"
asc version set-build --app <APP> --build 42          # ビルドはXcode/Transporterでアップロード済みが前提

# スクリーンショット（reserve→バイトPUT→commit を一括実行。asc api では不可能な処理）
asc assets upload-screenshot --app <APP> --locale ja --display APP_IPHONE_67 \
  --file 01.png --file 02.png
asc assets list --app <APP> --locale ja

# App内課金（ローカライズ / 価格 / 審査スクショ / 提出）
asc iap list --app <APP>
asc iap localize --app <APP> --product <productId> --name "..." --description "..."
asc iap price --app <APP> --product <productId> --territory JPN --price 150
asc iap screenshot --app <APP> --product <productId> --file iap-review.png

# App Review連絡先・メモ
asc review-detail set --app <APP> --first 邦彦 --last 宮永 --email contact@example.com --notes @notes.txt

# 審査へ提出（reviewSubmissions）
asc submit --app <APP>                # --prepare-only で最終送信せずステージのみ
```

### APIでは操作できず人間が行う必要がある工程

- **ビルド（.ipa）のアップロード**: App Store Connect APIに存在しません。Xcode / Transporter / `xcrun altool` で行い、`asc version set-build` で選択します。
- **「Appのプライバシー」データ収集ラベル**: 公開APIに書き込み口がなく、Web UIで入力します。
- **有料App契約・税・銀行情報（Paid Apps Agreement）**: Web UIで受諾します（未了だとIAPが有効になりません）。

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
