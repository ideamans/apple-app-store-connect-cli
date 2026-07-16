# 実利用フィードバック（asc CLI）

実アプリ「日本領収書スキャン」の**初回App Store提出一式**を、App Store Connect APIで通した際に得た知見。
CLIが「検証・自動補正・警告・ドキュメント化」すると事故が減る点を、優先度順にまとめる。
（作業自体はNode製の即席スクリプトで実施したが、対応する `asc` コマンドを併記する）

---

## 🔴 高優先（実際に時間を溶かした）

### 1. アセットアップロードは「commit 200」＝成功ではない。assetDeliveryState をポーリングすべき
- 症状: `inAppPurchaseAppStoreReviewScreenshots`（および `appScreenshots`）は
  **reserve→PUT→PATCH(uploaded:true) がすべて 2xx でも、その後の非同期処理で `FAILED` になる**ことがある。
  失敗時 `assetDeliveryState.errors[].code`（例 `IMAGE_INCORRECT_DIMENSIONS`）が付き、`imageAsset` は `{width:0,height:0}` になる。
- CLIの現状: `uploadAsset`（`cmd/assets.go`）は PATCH 後にIDを返して終了し、**処理結果を検証していない**。
  そのため「アップロード成功」と表示されても、実際は `FAILED` で反映されていないことがある（今回まさにこれで何度も踏んだ）。
- 提案: `uploadAsset` の後に `assetDeliveryState.state` を数秒間隔でポーリングし、
  `COMPLETE` を待つ／`FAILED` なら `errors[].code` を添えて非ゼロ終了。`asc assets upload-screenshot` と `asc iap screenshot` の両方で。

### 2. IAP審査用スクショの受理サイズは App Store 用と別。新サイズは全滅
- 実測（1商品で総当り）:
  - ✅ `1242×2208`（5.5"）→ COMPLETE
  - ❌ `1290×2796` / `1320×2868` / `1284×2778` / `1170×2532` / `1080×1920` / `993×2160` / `1206×2622`（元画像）
    → すべて `IMAGE_INCORRECT_DIMENSIONS`
- つまり **`appScreenshots`（App Store用）は新デバイスサイズを受理するが、`inAppPurchaseAppStoreReviewScreenshots` は
  レガシーな固定サイズ（1242×2208 等）しか受けない**。バリデータが別物。
- 提案: `asc iap screenshot` で
  - アップロード前に寸法を検証し、非対応なら**受理サイズ一覧を添えて即エラー**（無駄なアップロード往復を防ぐ）。
  - もしくは `--auto-fit` 的オプションで **アスペクト比を保ったまま 1242×2208 にリサイズ＋パディング**して投入
    （今回は `sips --resampleHeight 2208` → `sips -p 2208 1242 --padColor FFFFFF` で解決）。
  - 成功で IAP の state が `MISSING_METADATA` → `READY_TO_SUBMIT` に変わる。これを検知して表示できると親切。

---

## 🟡 中優先（ハマりどころ・非致命だがUXを損なう）

### 3. 初回バージョンでは `whatsNew` が編集不可
- `appStoreVersionLocalizations` の `whatsNew` を初回バージョンでPATCHすると
  `409 / "Attribute 'whatsNew' cannot be edited at this time"`。初回リリースにリリースノートは不要という仕様。
- 提案: `asc version localize --whats-new ...` は、初回バージョン時にこのエラーを**非致命として握り**、
  他の属性（description/keywords 等）の更新は成功させる（whatsNewだけスキップして警告）。全体を失敗にしない。

### 4. IAPは既に存在していることが多い（Sandboxテスト用に作成済み）→ create が 409
- `POST /v2/inAppPurchases` が `409 "This product ID has already been used"`。
- 提案: `asc iap` 系は **productId で find-or-create**。`iap localize/price/screenshot` は既存商品にそのまま効くべき
  （`resolveIAP` があるので概ね対応済みと思われるが、`iap create` の409は「既存を返す」挙動にすると事故が減る）。

### 5. 新しい年齢制限フローは GET で読めない（404）
- `GET /v1/appStoreVersions/{id}/ageRatingDeclaration` が **404**。UIでは 4+ が設定済みでも読めない。
- 提案: `asc appinfo age-rating` の read/verify は 404 を「未対応/読取不可」として扱い、
  「設定はUIで確認」と案内。設定(PATCH @attrs.json)側は動く前提でOK。

### 6. 管理APIには「App Store Connect APIキー（ロール付き）」が必須
- 同じ発行元でも、**App Store Server API 用（In-App Purchaseキー）は管理API（api.appstoreconnect.apple.com）で 401**。
  管理には **Admin/App Manager ロールの ASC API キー**が要る。
- 提案: `asc configure` 時、または 401 時に「Server APIキーでは管理APIは使えません。ロール付きのASC APIキーを使ってください」と誘導。

### 7. ES256 JWS 署名は raw(P1363, 64byte) で。DERだと 401
- 実装ノート（トークン生成）。Goの `ecdsa.Sign` は r,s を返すので、**各32byte固定長ゼロ埋め→連結**してから base64url。
  ASN.1 DER のまま入れると Apple は `401 NOT_AUTHORIZED`。既に対応済みなら無視でよいが、最頻出の実装バグなので回帰テストに。
  （JS実装で Buffer を JSON.stringify して base64url する類似バグも踏みやすい。）

### 8. インライン作成の一時idは `${local-id}` 形式（priceSchedules / availabilities）
- `appPriceSchedules` / `inAppPurchasePriceSchedules` / `appAvailabilities` を POST するとき、`included` の
  一時idを普通の文字列（例 `"p1"`）にすると **409 `ENTITY_ERROR.INCLUDED.INVALID_ID`**
  （"the id must be a local id with the format '${local-id}'"）。**`"${p1}"` のように `${...}` で囲む**必要がある。
- 提案: `asc pricing` / `asc availability` 系の内部でこの形式を使う（外から意識させない）。

### 9. 配信地域（appAvailabilities v2）は全テリトリーを明示する必要がある
- 「日本だけ available:true」で POST すると **409**：「territoryAvailabilities.territory は id 'SVN' の
  included を期待するが無い」等、**全テリトリー分の territoryAvailabilities を要求**される。
- つまり「日本のみ配信」でも **`/v1/territories`（約175件）を全取得 → JPN=true, 他=false を全部 included** で送る。
- 提案: `asc availability set --territory JPN [--only]` で、CLI が全テリトリーを取得して差分を組み立てる。

### 10. 無料アプリの価格設定は customerPrice "0" の price point を探す
- 無料は `appPriceSchedules` に **customerPrice が "0" の appPricePoint** を baseTerritory 付きで紐付ける
  （`/v1/apps/{id}/appPricePoints?filter[territory]=JPN` を走査して探す）。ページング必須（1ページ200件）。
- 提案: `asc pricing set --free`（または `--price 0`）で、この price point 解決を隠蔽する。
- 注意: 価格・配信地域は**提出の必須項目なのに、バージョン画面が完成して見えても未設定のまま**になりがち
  （UIでも「価格を追加」「配信状況の設定」ボタンが残る）。`asc submit` の事前チェックに含めると親切。

---

## 🟢 「APIでは操作できず人間が行う」に追記したい項目
（READMEの該当セクションの補強）

- **App Store Server Notifications V2 の URL 設定は GUI 専用**（App Store Connect API にも Server API にも設定エンドポイントなし）。
  疎通は Server API の「Request a Test Notification」で確認できる（配信結果は sendAttempts の statusCode）。
- **App のプライバシー（データ収集ラベル / nutrition labels）は公開APIに存在しない**（確認済み：
  `/v1/apps/{id}/dataUsages`・`appDataUsages`・`dataUsageGroupings` すべて **404**）。**GUI専用**。
  `asc` では「非対応」と明記し、`asc submit` 前チェックでは触れない（人手で公開済みか案内する程度）。
  参考: GA4/Firebase Analytics の正しい申告は **使用状況データ>製品の操作 / 用途=分析 / 非紐付け / 非トラッキング**
  （「広告データ」ではない）。
- **日本の特定商取引法に基づく表記** は ASC に専用欄が無い（App情報にあるのは EU DSA「デジタルサービス法 トレーダーステータス」で、
  これはEU配信向け）。日本のみ配信なら Web の特商法ページで足りる、という整理を明記すると混乱が減る。

---

## 参考: 今回通した工程（CLIコマンド対応）
- appinfo: 名称/サブタイトル/プライバシーURL・カテゴリ → `asc appinfo localize/category`
- version: 説明/キーワード/プロモ/サポートURL・著作権・ビルド紐付け → `asc version localize/set-build`（copyrightはversion属性PATCH）
- スクショ: iPhone/iPad → `asc assets upload-screenshot`（**#1 のポーリング要**）
- IAP: 表記/価格/審査スクショ（4商品）→ `asc iap localize/price/screenshot`（**#2 のサイズ要**）
- 審査情報: 連絡先/メモ → `asc review-detail set`
- 提出: `asc submit`（プライバシーURL公開・IAP同梱が前提）
