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
asc version set-build --app <APP> --build 42          # ビルドはXcode/Transporterまたは asc builds upload でアップロード済みが前提

# スクリーンショット（reserve→バイトPUT→commit を一括実行。asc api では不可能な処理）
asc assets upload-screenshot --app <APP> --locale ja --display APP_IPHONE_67 \
  --file 01.png --file 02.png
asc assets list --app <APP> --locale ja

# App内課金（ローカライズ / 価格 / 審査スクショ / 提出）
asc iap list --app <APP>
asc iap localize --app <APP> --product <productId> --name "..." --description "..."
asc iap price --app <APP> --product <productId> --territory JPN --price 150
asc iap screenshot --app <APP> --product <productId> --file iap-review.png --auto-fit  # 審査スクショはレガシーサイズのみ受理。--auto-fitで自動リサイズ

# App Review連絡先・メモ
asc review-detail set --app <APP> --first 邦彦 --last 宮永 --email contact@example.com --notes @notes.txt

# 審査へ提出（reviewSubmissions）
asc submit --app <APP>                # --prepare-only で最終送信せずステージのみ
```

## 全コマンド一覧（APIドメイン別）

App Store Connect API（OpenAPI 4.4.1）の全ドメインを専用サブコマンドでカバーしています。書き込み系はすべて `--dry-run` 対応。詳細は `asc <command> --help` / `asc --llm` を参照。

| コマンド | 対応ドメイン |
|---------|-------------|
| `apps` | アプリ一覧・詳細・更新（primaryLocale、コンテンツ権利） |
| `appinfo` | App情報（名前・サブタイトル・カテゴリ・年齢レーティング） |
| `version` | バージョン（作成・更新・削除・ローカライズ・ビルド選択・リリース要求・段階的リリース） |
| `assets` | スクリーンショット / プレビュー動画 / ルーティングカバレッジのアップロード・管理 |
| `review-detail` | 審査連絡先・デモアカウント・メモ・添付ファイル |
| `submit` / `review-submissions` | 審査提出、提出の一覧・キャンセル・アイテム管理 |
| `pricing` / `availability` / `territories` | アプリ価格スケジュール・価格ポイント・配信地域・予約注文終了 |
| `iap` | App内課金（作成・削除・ローカライズ・価格・配信可否・プロモ画像・オファーコード・審査提出） |
| `subscriptions` | サブスクリプション（グループ・価格・イントロ/プロモオファー・オファーコード・配信可否・グレースピリオド・プロモート表示・プロモ画像・ドラフト版・審査提出） |
| `beta` / `sandbox` | TestFlight（ベータグループ・テスター・ローカライズ・ベータ審査・ライセンス・自動募集条件・App Clip起動・ビルド詳細）、サンドボックステスター |
| `builds` | ビルド（一覧・失効・暗号化フラグ・テスター通知・**IPAアップロード**） |
| `background-assets` | Background Assets（アセットパックの作成・バージョン・アップロード） |
| `users` / `invitations` | チームユーザー・招待 |
| `certificates` / `bundle-ids` / `devices` / `provisioning-profiles` / `pass-type-ids` / `merchant-ids` | 署名・プロビジョニング一式 |
| `reports` / `analytics` | セールス・ファイナンスレポート、アナリティクスレポートの要求とダウンロード |
| `reviews` | カスタマーレビューと開発者返信 |
| `metrics` / `diagnostics` / `feedback` | パフォーマンス・電力メトリクス、診断ログ、TestFlightフィードバック |
| `events` | In-App Event（作成・スケジュール・ローカライズ・画像/動画） |
| `product-pages` / `experiments` | カスタム製品ページ、製品ページ最適化（A/Bテスト） |
| `nominations` / `eula` | フィーチャリングノミネーション、カスタムEULA |
| `gamecenter` | Game Center（実績・リーダーボード・セット・グループ・アクティビティ・チャレンジ・マッチメイキング・バージョン単位の有効化・スコア/実績のサーバー投稿） |
| `appclips` | App Clip（体験・ローカライズ・ヘッダー画像・審査情報） |
| `xcode-cloud` | Xcode Cloud（プロダクト・ワークフロー・ビルド実行・アーティファクト） |
| `webhooks` | Webhook（作成・配信履歴・ping） |
| `encryption` / `accessibility` | 輸出コンプライアンス（暗号化申告）、アクセシビリティ宣言 |
| `alt-distribution` / `marketplace` / `app-tags` / `actors` / `android-mapping` | EU代替配布・マーケットプレイス・Appタグ・操作主体の参照・Android→iOS対応付け |

専用コマンド化していない枝葉（win-backオファー作成、App Clip advanced experience作成など、公開スペックが不完全なもの）は `asc api` で直接呼び出せます。全実装はOpenAPIスペック4.4.1と突き合わせてレビュー済みで、deprecatedエンドポイントは使用していません（Apple側に代替がないmarketplace webhookを除く）。

### APIでは操作できず人間が行う必要がある工程

- **「Appのプライバシー」データ収集ラベル**: 公開APIに存在しません（`dataUsages`系エンドポイントは404）。Web UIで入力します。参考: GA4/Firebase Analyticsの申告は「使用状況データ > 製品の操作 / 用途=分析 / 非紐付け / 非トラッキング」（「広告データ」ではない）。
- **有料App契約・税・銀行情報（Paid Apps Agreement）**: Web UIで受諾します（未了だとIAPが有効になりません）。
- **App Store Server Notifications V2のURL設定**: 設定エンドポイントがGUI専用（疎通はServer APIのTest Notificationで確認可能）。
- **日本の特定商取引法に基づく表記**: ASCに専用欄はありません（App情報の「トレーダーステータス」はEU DSA向け）。日本のみ配信ならWebの特商法ページで対応します。

※ ビルド（.ipa）のアップロードはAPI 4.xの `buildUploads` に対応した `asc builds upload` で可能になりました（従来どおりXcode / Transporterも利用可）。

### 実提出で判明した落とし穴への対策（組み込み済み）

- **アセットの非同期検証**: commit 2xx後もAppleが非同期で検証し、寸法エラー等は後から`assetDeliveryState`に出ます。画像・動画アセットの全アップロードコマンドがCOMPLETE/FAILEDまでポーリングし、失敗時はエラーコード付きで非ゼロ終了します（ビルド・Background Assetsの処理状況は `builds list` / `background-assets versions` で確認）。
- **IAP/サブスク審査スクショはレガシーサイズのみ受理**（1242×2208 / 2208×1242 / 2048×2732 / 2732×2048）。`iap screenshot`/`subscriptions screenshot`はアップロード前に寸法検証し、`--auto-fit`でアスペクト比維持リサイズ+白パディングを行えます。
- **初回バージョンのwhatsNewは編集不可**（409）→ `version localize`は警告してスキップし他の属性を適用。
- **`iap create`はfind-or-create**（Sandboxテスト用に作成済みでも409にならない）。
- **配信地域は全テリトリー明示が必要** → `availability set`が約175テリトリーを自動展開（指定外はfalse）。
- **無料アプリにも価格スケジュールが必要** → `pricing set --free`。`asc submit`は価格・配信地域の未設定を事前警告します。

## ヘルプ

- `asc --help` / `asc <command> --help` — 人間向けの通常のヘルプ

### AIエージェントから使う

`asc llm` はAIエージェント向けの詳細リファレンスを一括出力します（クレデンシャルモデル、解決順序、**実提出で判明した落とし穴**、全コマンド・フラグ、`asc api` のJSON:API/ページネーションの注意点）。コマンドカタログはcobraのコマンド定義から生成されるため実装と乖離せず、バイナリに埋め込まれているのでオフラインでも実行中のバージョンと必ず一致します。

```bash
asc llm                  # Markdown（約5,700行。grepして使う想定）
asc llm --format json    # 章ごとのJSON配列
asc --llm                # 非推奨エイリアス。従来どおりどのサブコマンドに付けても動作します
```

Claude Code ではプラグインを導入すると `/asc-usage` と `/asc-install` が使えます。

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install apple-app-store-connect-cli@ideamans-plugins
```

同じスキルは Copilot や Cursor など Agent Skills 対応ホストでも利用できます。

```bash
gh skill install ideamans/apple-app-store-connect-cli/plugins/apple-app-store-connect-cli/skills/asc-usage --agent copilot
```

スキル本体は [`plugins/apple-app-store-connect-cli/`](plugins/apple-app-store-connect-cli)、準拠している標準は [ideamans/go-llm-cli-kit](https://github.com/ideamans/go-llm-cli-kit) を参照してください。

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
