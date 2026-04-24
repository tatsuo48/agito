# Figaro — MVP 仕様書（v0.1）

ローカル LLM（Ollama）を使って、git の変更内容からブランチ作成 → コミット → PR 作成までを一気通貫で自動化する CLI ツール。

本書は Claude Code が実装するための仕様・設計書。実装に着手する前に、本書の「前提条件」「ワークフロー」「Ollama 連携」セクションを読み込むこと。

---

## 1. コンセプト

作業ツリーにある変更内容を見て、Figaro が以下を全部やってくれる：

1. 適切なブランチ名を考えて、ブランチを作成・切り替え
2. 適切な commit message を考えて、commit
3. PR の title / description を考えて、PR を作成

LLM 推論はすべてローカルの Ollama に投げる。外部 API は使わない（GitHub API は `gh` CLI 経由なので例外）。

設計の肝は **「速くて・オフラインで動いて・コードが外に出ない」**。Sherlock（クラウド・高性能）と対照的な位置付け。

---

## 2. MVP スコープ

### やること（v0.1）

- `figaro` 1 コマンドで、変更検知 → ブランチ作成 → コミット → push → PR 作成までの一気通貫フロー
- デフォルトモデルは `gemma4:latest`（`--model` で上書き可能）
- 日本語の commit message / PR description を生成
- 実行前に生成結果を人間が確認できる（y / n / edit / regenerate）

### やらないこと（v0.1 スコープ外）

- 複数コミットへの分割
- Issue との自動紐付け
- レビュアー自動指定
- テンプレート（`.figaro.yml`）対応
- コンフリクト解消
- `figaro commit` / `figaro pr` など個別サブコマンド（将来追加）

---

## 3. コマンドインターフェース

```
figaro [options]
```

| オプション      | デフォルト               | 説明                                               |
| --------------- | ------------------------ | -------------------------------------------------- |
| `--model`       | `gemma4:latest`          | 使用する Ollama モデル                             |
| `--ollama-host` | `http://localhost:11434` | Ollama エンドポイント                              |
| `--language`    | `ja`                     | 生成言語（`ja` or `en`）                           |
| `--dry-run`     | `false`                  | 生成結果の表示のみ。実際の git / gh 操作は行わない |
| `--yes` / `-y`  | `false`                  | 確認プロンプトをスキップ（非推奨、CI 用）          |
| `--draft`       | `false`                  | PR を draft で作成                                 |
| `--base`        | 自動検出                 | PR のベースブランチ                                |
| `--no-pull`     | `false`                  | デフォルトブランチへ移動後の `git pull` をスキップ |

終了コード：成功 `0` / ユーザーキャンセル `1` / git 操作失敗 `2` / Ollama 失敗 `3` / gh 失敗 `4` / 前提条件不足 `10`

---

## 4. 前提条件

実行時に以下をチェックし、欠けていれば終了コード `10` で終了（明確なエラーメッセージ付き）：

- `git` がインストール済み、カレントディレクトリが git リポジトリ
- `gh` がインストール済み、認証済み（`gh auth status` で確認）
- `ollama serve` が起動中、指定モデルが `ollama list` に存在する
- リモート `origin` が設定済み（`git remote get-url origin`）
- 何らかの変更が存在する（tracked の変更 or untracked ファイル）

---

## 5. メインワークフロー

Figaro は、スタート時点でユーザーがどのブランチにいるかを問わない。まずデフォルトブランチに移動し、そこを起点として新しいブランチを切る。現在の変更内容は stash 経由で新ブランチに持ち込む。

```
┌──────────────────────────────────────────┐
│ 1. 前提条件チェック                       │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 2. 変更検出                               │
│    staged + unstaged + untracked を検出  │
│    変更なし → 終了                        │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 3. デフォルトブランチ検出                 │
│    origin/HEAD → main / master / ...     │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 4. デフォルトブランチへ移動               │
│    a. 現在の変更を stash                  │
│       (--include-untracked)               │
│    b. git checkout {default_branch}       │
│    c. git pull --ff-only (--no-pull で抑止)│
│    d. git stash pop                       │
│                                           │
│  ※ カレント == デフォルトブランチなら    │
│    stash/pop は省略（pull だけ行う）     │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 5. diff 取得                              │
│    git diff HEAD + untracked の疑似 diff │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 6. Ollama で生成                          │
│    branch_name / commit / pr_title /     │
│    pr_body を 1 回の推論で JSON 取得      │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 7. ユーザー確認                           │
│    [y]es [e]dit [r]egenerate [n]o       │
└──────────────────────────────────────────┘
                  │
                  ▼
┌──────────────────────────────────────────┐
│ 8. 実行                                   │
│    ├─ git checkout -b {branch_name}      │
│    ├─ git add -A                          │
│    ├─ git commit -F -                     │
│    ├─ git push -u origin HEAD            │
│    └─ gh pr create --title --body-file   │
└──────────────────────────────────────────┘
                  │
                  ▼
              PR URL を表示
```

**重要**：ステップ 4 で stash pop に失敗した場合（デフォルトブランチの pull で取り込まれた変更と衝突した場合など）、Figaro は自動解決を試みず、stash を残したままエラー終了する。ユーザーが解消後に再実行する想定。詳細は §8。

---

## 6. Ollama 連携仕様

### 6.1. エンドポイント

`POST {ollama-host}/api/generate`

```json
{
  "model": "gemma4:latest",
  "prompt": "...",
  "format": "json",
  "stream": false,
  "options": {
    "temperature": 0.3,
    "num_ctx": 16384
  }
}
```

`format: "json"` を必ず指定し、JSON 出力を強制する。`temperature` は低め（0.2〜0.4）で揺らぎを抑える。

### 6.2. プロンプト

システムプロンプト相当の指示 + diff を渡す。プロンプトは以下の構造：

```
You are Figaro, a git assistant that generates branch names, commit messages, and pull requests from git diffs.

Analyze the following git diff and produce a JSON object with exactly these fields:

- "branch_name": kebab-case branch name prefixed with one of: feat/, fix/, chore/, docs/, refactor/, test/, perf/. Max 50 chars. English only.
- "commit_type": one of: feat, fix, chore, docs, refactor, test, perf
- "commit_scope": short scope in English (e.g., "api", "cache", "auth") or null
- "commit_subject": imperative mood, max 72 chars. Written in {language}.
- "commit_body": markdown explaining what changed and why. Written in {language}. 2-5 sentences.
- "pr_title": similar to commit_subject but can include conventional commit prefix. Written in {language}.
- "pr_body": markdown with sections "## 背景", "## 変更内容", "## 確認方法" (if language=ja) or "## Context", "## Changes", "## How to verify" (if language=en).

Rules:
- Output ONLY a JSON object. No markdown code fences. No preamble.
- Do not include file paths verbatim in the subject line.
- If the diff is trivial (typo, whitespace), keep the messages short.
- If you cannot determine intent from the diff, use "chore" as the type.

<diff>
{diff_content}
</diff>
```

`{language}` と `{diff_content}` はランタイムで埋め込む。

### 6.3. 期待する JSON スキーマ

```json
{
  "branch_name": "fix/elasticache-ttl-policy",
  "commit_type": "fix",
  "commit_scope": "cache",
  "commit_subject": "ElastiCache の maxmemory-policy を allkeys-lru に変更",
  "commit_body": "TTL が設定されていないキーでメモリが枯渇する問題を修正。\n\n...",
  "pr_title": "fix(cache): ElastiCache の maxmemory-policy を allkeys-lru に変更",
  "pr_body": "## 背景\n\n...\n\n## 変更内容\n\n...\n\n## 確認方法\n\n..."
}
```

### 6.4. 最終的な commit message の組み立て

LLM の出力をそのまま使わず、以下のテンプレートでレンダリングする：

```
{commit_type}({commit_scope}): {commit_subject}

{commit_body}
```

`commit_scope` が `null` の場合は `{commit_type}: {commit_subject}` にする。

### 6.5. バリデーション

LLM 出力を受け取ったら、以下を検証：

- 必須フィールドが全て存在するか
- `branch_name` が `^(feat|fix|chore|docs|refactor|test|perf)/[a-z0-9-]+$` にマッチするか
- `branch_name` が 50 文字以内か
- `commit_type` が enum に含まれるか
- `commit_subject` が 72 文字以内か（日本語の場合は文字数基準。バイト数ではない）

バリデーション失敗時は最大 2 回までリトライ（プロンプトに「Previous output was invalid because X」を追記して再実行）。

---

## 7. ステップ詳細

### 7.1. デフォルトブランチ検出

```bash
git symbolic-ref refs/remotes/origin/HEAD --short
# → "origin/main"
```

これが失敗する場合のフォールバック：

```bash
git remote show origin | grep "HEAD branch" | awk '{print $NF}'
```

さらに失敗したら `main` → `master` → `develop` の順で存在確認。全部無かったらエラー終了。

### 7.2. デフォルトブランチへの移動

スタート時のカレントブランチによらず、一度デフォルトブランチに移動する。

#### a. カレントが **デフォルトブランチと同じ** 場合

```bash
git pull --ff-only origin {default_branch}   # --no-pull 指定時はスキップ
```

stash は不要。そのまま §7.3 の diff 取得へ。

#### b. カレントが **デフォルトブランチと異なる** 場合

```bash
# 1. untracked も含めて stash
git stash push --include-untracked -m "figaro-auto-stash-{timestamp}"

# 2. 変更が何もなくて stash が作られなかった場合は終了（§5 step 2 で弾くので通常来ない）

# 3. デフォルトブランチへ移動
git checkout {default_branch}

# 4. 最新化
git pull --ff-only origin {default_branch}   # --no-pull 指定時はスキップ

# 5. stash を戻す
git stash pop
```

**失敗時の挙動**：

| 失敗箇所                   | 対応                                                                                                                                 |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `stash push` 失敗          | 即エラー終了。git の状態はそのまま                                                                                                   |
| `checkout` 失敗            | stash を pop で戻して終了（ユーザーの元ブランチに残すイメージ）                                                                      |
| `pull --ff-only` 失敗      | 「デフォルトブランチがローカルで進んでいます。手動で解消してください」とメッセージ。stash を pop して終了                            |
| `stash pop` でコンフリクト | stash は **残したまま** 終了。「コンフリクトを解消して `git stash pop` してから Figaro を再実行してください」と案内。stash ID を明示 |

`--no-pull` オプション指定時は pull をスキップするので、ネットワーク切断時やオフライン作業時でも動く。

### 7.3. diff 取得

対象：

1. tracked ファイルの変更（staged + unstaged）: `git diff HEAD`
2. untracked ファイル: `git ls-files --others --exclude-standard` で列挙し、各ファイルを疑似 diff に整形

```
diff --git a/new_file.py b/new_file.py
new file mode 100644
+++ b/new_file.py
@@ -0,0 +1,N @@
+{ファイル内容}
```

#### トークン制限対応

diff が巨大な場合（概ね 30,000 文字を超える場合）は以下で縮約：

1. `lock` ファイル、`node_modules`, `vendor`, 自動生成ファイル（`*.pb.go`, `*.generated.ts` 等）は除外
2. それでも大きければ、ファイルごとの変更行数を集計して「変更の多い順」にトップ N ファイルだけ diff に含める。残りはファイル名と変更行数のみ伝える
3. それでも超過する場合は、各ファイルの diff を先頭 50 行 + 末尾 20 行に truncate

### 7.4. ブランチ作成

§7.2 を経た時点でカレントはデフォルトブランチ、かつ作業ツリーに変更が反映されている状態。ここから新ブランチを切る：

```bash
git checkout -b {branch_name}
```

`branch_name` が既に存在する場合は末尾に `-2`, `-3` を付けて衝突回避。

### 7.5. コミット

```bash
git add -A
git commit -F -   # stdin から commit message を渡す
```

### 7.6. push

```bash
git push -u origin HEAD
```

失敗した場合（権限なし、リモートが進んでいる等）はエラーメッセージを表示して終了。pull や rebase は自動で行わない（安全優先）。

### 7.7. PR 作成

```bash
gh pr create \
  --title {pr_title} \
  --body-file - \
  --base {default_branch} \
  [--draft]
```

body は stdin から渡す。成功したら PR URL を標準出力に出す。

`.github/pull_request_template.md` が存在しても、v0.1 では無視する（LLM 生成結果を優先）。将来は統合する。

---

## 8. エラーハンドリング

| ケース                                        | 挙動                                                                                                                         |
| --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Ollama がダウン                               | 接続エラーを明示し、`ollama serve` 実行を案内                                                                                |
| モデルが未 pull                               | `ollama pull {model}` を実行するか確認 → Yes なら pull                                                                       |
| LLM 出力が JSON として不正                    | 最大 2 回リトライ。それでもダメなら raw 出力を表示して終了                                                                   |
| バリデーションエラー                          | 「branch_name が規約違反」等の理由を添えてリトライ                                                                           |
| 変更が存在しない                              | 「No changes detected」と表示して終了コード 0                                                                                |
| stash push 失敗                               | 即エラー終了。git の状態は変えない                                                                                           |
| デフォルトブランチへの checkout 失敗          | stash を pop で戻してから終了                                                                                                |
| `git pull --ff-only` 失敗（non-fast-forward） | stash を pop で戻してから、「デフォルトブランチが競合しているため `--no-pull` で再実行するか、手動で解消してください」と案内 |
| stash pop でコンフリクト                      | stash を残したまま終了。stash ID を明示し、「コンフリクト解消 → `git stash drop` → Figaro 再実行」の手順を案内               |
| すでに同名ブランチ存在                        | サフィックス `-2`, `-3` を付けて再試行                                                                                       |
| push 失敗                                     | ブランチ作成・commit は残したまま、「git push -u origin HEAD を手動実行してください」と案内                                  |
| PR 作成失敗                                   | push まで完了している場合は「gh pr create ... で手動作成できます」と案内、body を一時ファイルに保存                          |

中断時は git 側の状態を「一貫した状態」に保つ。具体的には：

- stash 段階で失敗 → 元のブランチ・元の作業ツリー状態を維持
- stash pop コンフリクト → stash は残す（ユーザーが解消可能な状態）
- ブランチ作成成功 → commit 失敗なら、ブランチは残すが stage は戻さない（ユーザーが調査できる状態）
- commit 成功 → push 失敗なら、ローカルの commit は残す（再 push 可能）
- push 成功 → PR 作成失敗なら、push された状態は残す（手動で PR 作成可能）

stash を作った場合、成功時には必ず `git stash list` から Figaro が作った stash（メッセージ `figaro-auto-stash-*`）が消えていることを最終チェックする。残っていたら警告を出す。

---

## 9. ユーザー体験

### 9.1. 実行時の画面遷移（例）

```
$ figaro
✓ git, gh, ollama all available
✓ default branch: main
✓ current branch: wip/debug
✓ detected 3 changed files (+42 -18 lines)

Moving to default branch...
  ✓ stashed changes (figaro-auto-stash-20260424-143022)
  ✓ checked out main
  ✓ pulled latest (up to date)
  ✓ popped stash

Generating with gemma4:latest...
  [■■■■■■■■■■] done (2.3s)

─────────────────────────────────────────
  Branch:  fix/elasticache-ttl-policy
  Commit:  fix(cache): ElastiCache の maxmemory-policy を変更

  TTL が設定されていないキーでメモリが枯渇する問題を
  修正。allkeys-lru に切り替えることで...

  PR Title: fix(cache): ElastiCache の maxmemory-policy を変更
  PR Body:  (71 lines) — preview with 'p'
─────────────────────────────────────────

Proceed? [y]es / [e]dit / [r]egenerate / [p]review PR body / [n]o:
```

### 9.2. 編集モード（`e`）

以下を選択させて `$EDITOR` を起動：

- branch_name のみ編集
- commit message のみ編集
- PR title / body のみ編集
- 全部まとめて編集（YAML 形式）

編集後、再度サマリを表示して確認を取る。

### 9.3. 再生成（`r`）

オプションで追加指示を受け取れると嬉しい：

```
Regenerate with extra instruction (or empty to just retry):
> もう少し簡潔に。背景セクションは 2 文以内で。
```

プロンプトに `Additional user instruction: {...}` を追記して再実行。

---

## 10. 技術選定

### 推奨言語：Go

理由：

- 単一バイナリ配布（`brew install figaro` で完結）
- 起動速度が速い（CLI の体感 UX に直結）
- git / gh の薄いラッパーとして書きやすい
- ollama / HTTP クライアントのエコシステムが十分

ディレクトリ構成：

```
figaro/
├── cmd/figaro/main.go           # エントリポイント
├── internal/
│   ├── git/                     # git コマンドラッパー
│   │   ├── branch.go
│   │   ├── diff.go
│   │   └── commit.go
│   ├── gh/                      # gh CLI ラッパー
│   │   └── pr.go
│   ├── ollama/                  # Ollama クライアント
│   │   ├── client.go
│   │   └── prompt.go
│   ├── generate/                # 生成ロジック（バリデーション含む）
│   │   ├── schema.go
│   │   └── validate.go
│   ├── ui/                      # 対話 UI
│   │   ├── prompt.go
│   │   └── editor.go
│   └── flow/                    # メインワークフロー
│       └── flow.go
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

### 依存ライブラリ

- `net/http` （標準）— Ollama 通信
- `encoding/json` （標準）
- `os/exec` （標準）— git / gh 呼び出し
- `github.com/charmbracelet/lipgloss` — 端末装飾（オプション）
- `github.com/charmbracelet/huh` or `github.com/AlecAivazis/survey/v2` — 対話プロンプト

---

## 11. 実装の順序（推奨）

Claude Code で作るなら、以下の順で漸進的に実装すると詰まりにくい：

1. **前提条件チェック + デフォルトブランチ検出** だけ動くバイナリ
2. **diff 取得** を実装し、標準出力に diff を吐くだけのモードを追加
3. **Ollama 連携** を実装し、diff から JSON を返せるようにする（git 操作はまだしない）
4. **バリデーション** を実装
5. **デフォルトブランチへの移動**（stash → checkout → pull → stash pop）を実装。ここは異常系が多いので、`--dry-run` で各ステップの動作を確認しながら進めること。stash pop 失敗時の後処理を特に丁寧に
6. **新ブランチ作成 + commit + push** を実装
7. **gh 操作**（PR 作成）を実装
8. **対話 UI**（確認・編集・再生成）を実装
9. エラーハンドリングの網羅
10. README 書く

各段階で `--dry-run` が効くようにしておくと、デバッグが楽。

---

## 12. テスト観点

- 変更なしでの起動 → 終了コード 0、メッセージのみ
- tracked 変更のみ（staged / unstaged 混在）
- untracked ファイル追加
- tracked + untracked 混在
- 巨大 diff（100 ファイル超、10万行超）→ truncate が効くか
- 日本語のみの変更 / 英語のみ / 混在
- バイナリファイルが含まれる diff
- **スタート時のブランチ状態のバリエーション**
  - カレント == デフォルトブランチ（stash 不要パス）
  - カレント != デフォルトブランチ（stash → checkout → pop パス）
  - カレント != デフォルトブランチ、かつデフォルトブランチがリモートで進んでいる（pull 必要）
  - カレント != デフォルトブランチ、`--no-pull` 指定
  - stash pop でコンフリクトが起きるケース（異常系）
- デフォルトブランチが `main` / `master` / `develop` のケース
- 既に同名ブランチが存在するケース
- Ollama 停止中
- gh 未認証
- LLM が JSON 以外を返すケース
- ネットワーク不通（push / gh 段階）

---

## 13. 将来の拡張（v0.1 ではやらない）

- `.figaro.yml` でチーム規約を注入（scope 候補、PR テンプレート、言語など）
- `figaro commit` / `figaro pr` の個別サブコマンド
- 過去の commit 履歴を少量 RAG してチームのスタイルに寄せる
- 複数モデルでの一次生成 + 校正パイプライン
- Issue 自動紐付け（`Close #123` の付与）
- prepare-commit-msg フック統合
- Sherlock との接続（MCP サーバーとして振る舞うモード）

---

## 14. Claude Code への引き継ぎメモ

- 本書の「5. メインワークフロー」「6. Ollama 連携」「7. ステップ詳細」「10. 技術選定」が実装の中核。この 4 セクションは変更せずに実装してほしい
- 「11. 実装の順序」に従って段階的に動かしながら進めると、各段階で手触りを確認できる
- 対話 UI のライブラリ選定（huh vs survey）は Claude Code の好みで決めてよい
- README は最後に書く。内容は「インストール手順 / 使い方 / トラブルシューティング / 設定」の 4 セクション
- テストは table-driven test を基本に、`internal/generate/validate_test.go` と `internal/git/diff_test.go` を最低限用意する
