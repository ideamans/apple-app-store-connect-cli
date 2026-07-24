# CLAUDE.md — apple-app-store-connect-cli

Apple App Store Connect API の非対話 CLI。**バイナリ名は `asc`**、goreleaser の
プロジェクト名は `apple-app-store-connect`（`-cli` なし）、リポジトリ名は
`apple-app-store-connect-cli`。**3つとも違う**のでリリースアセット名やインストール
手順を書くときは取り違えないこと。

## 変更時の必須手順

**機能を追加した、フラグを増やした、既存の挙動を変えた — このいずれかをしたら、
3か所すべてを更新してから終わること。**

| 更新先 | 対象 | やり方 |
| --- | --- | --- |
| ① ドキュメント | `README.md` | 手で更新。使い方が変わったときのみ |
| ② ヘルプ | cobra の `Short` / `Long` / `Example` / フラグ説明 | コード内。**カタログはここから生成される**ので、ここを厚くすると③も良くなる |
| ③ **LLMナレッジ** | `internal/llmdocs/00-guide.md` | 認証・`asc api` の使い方・代表フローが変わったら |
| | `internal/llmdocs/10-pitfalls.md` | **実提出で判明した罠を見つけたら必ず追記** |
| | `internal/llmdocs/90-commands.md` | **生成物。手編集しない** → `go generate ./...` |
| | `plugins/apple-app-store-connect-cli/skills/*/SKILL.md` | 手順や前提が変わったとき |
| | `context7.json` の `rules` | 新しい落とし穴が生まれたとき |

③ を忘れやすい。ドキュメントとヘルプは人間が読んで気づくが、**LLMナレッジが
古いことには誰も気づかない**（エージェントが黙って間違えるだけ）。

判断に迷ったときの目安:

- **App Store の実提出で罠を踏んだ** → `10-pitfalls.md` に追記する。この章が
  この CLI で最も価値のある資産で、API ドキュメントにも書かれていない知識が
  溜まっている。`context7.json` の `rules` にも要約を入れる
- 新しいコマンドを足した → ②の `Short` / `Long` / `Example` を書いてから
  `go generate ./...`
- 認証やプロファイルの扱いを変えた → `00-guide.md` の「Credential model」と
  `context7.json`
- 破壊的・顧客可視な操作（submit / pricing / availability）を追加した →
  `asc-usage` の SKILL.md に確認手順を保つこと

## リリース

`PluginVersion`（`cmd/root.go`）と `plugin.json` の `version` と git タグの3つを
揃える。テストとリリースワークフローが不一致を検出する。手順は
`plugins/apple-app-store-connect-cli/PUBLISH.md`。

## 秘密情報

`.p8` 鍵は `.gitignore` 済み（`*.p8`）。**内容を出力しない・コマンドラインに
書かない・ユーザーのコピーを移動や削除しない**。Apple から再ダウンロードできない。

## 確認

```bash
go generate ./...     # 生成物を作り直す
git diff --exit-code  # 差分が出たらコミット漏れ
go test ./...         # SKILL.md 検証とバージョン整合を含む
go run ./cmd/asc llm | head
```

## 参照

- 標準: <https://github.com/ideamans/go-llm-cli-kit/blob/main/LLM.md>
- 生成物と原本の対応: `.claude/rules/ai-artifacts-policy.md`
- 再生成: `/regen-ai`
