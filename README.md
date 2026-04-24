こんにちは。これはテストです。

# Figaro

ローカルLLM（Ollama）を使って、gitの変更内容からブランチ作成 → コミット → PR作成までを一気通貫で自動化するCLIツール。

## 特徴

- **オフライン動作** — LLM推論はすべてローカルのOllamaに投げる。コードが外に出ない
- **高速** — Goのシングルバイナリで起動オーバーヘッドなし
- **安全** — 実行前に生成結果を確認・編集・再生成できる

## インストール

### Go でインストール

```bash
go install github.com/moneyforward/figaro/cmd/figaro@latest
```

### ソースからビルド

```bash
git clone https://github.com/moneyforward/figaro
cd figaro
make install
```

## 前提条件

- `git` がインストール済みで、カレントディレクトリがgitリポジトリ
- `gh`（GitHub CLI）がインストール済みで認証済み（`gh auth login`）
- `ollama serve` が起動中
- デフォルトモデル `gemma4:latest` がpull済み（`ollama pull gemma4:latest`）

## 使い方

```bash
figaro [options]
```

作業ツリーの変更を検知し、以下を自動実行します：

1. ブランチ名を生成してブランチを作成
2. コミットメッセージを生成してコミット
3. PR title / description を生成してPRを作成

### オプション

| オプション      | デフォルト               | 説明                                           |
| --------------- | ------------------------ | ---------------------------------------------- |
| `--model`       | `gemma4:latest`          | 使用するOllamaモデル                           |
| `--ollama-host` | `http://localhost:11434` | Ollamaエンドポイント                           |
| `--dry-run`     | false                    | 生成結果の表示のみ。実際のgit/gh操作は行わない |
| `--yes` / `-y`  | false                    | 確認プロンプトをスキップ（CI用）               |
| `--draft`       | false                    | PRをdraftで作成                                |
| `--base`        | 自動検出                 | PRのベースブランチ                             |
| `--no-pull`     | false                    | デフォルトブランチへの`git pull`をスキップ     |

### 終了コード

| コード | 意味               |
| ------ | ------------------ |
| 0      | 成功               |
| 1      | ユーザーキャンセル |
| 2      | git操作失敗        |
| 3      | Ollama失敗         |
| 4      | gh失敗             |
| 10     | 前提条件不足       |

## トラブルシューティング

### Ollama が起動していない

```bash
ollama serve
```

### モデルが存在しない

```bash
ollama pull gemma4:latest
```

### `stash pop` でコンフリクト

Figaroはコンフリクト時に自動解消を試みず、stashを残したまま終了します。
表示されたstash IDを確認して手動で解消してください：

```bash
# コンフリクトの確認
git status

# コンフリクト解消後
git add .
git stash drop stash@{0}

# Figaro を再実行
figaro
```

### push失敗後の手動実行

```bash
git push -u origin HEAD
```

### PR作成失敗後の手動実行

```bash
# Figaro が表示するコマンドを実行
gh pr create --title "..." --body-file /tmp/figaro-pr-body-*.md --base main
```

## 開発

```bash
# テスト実行
make test

# ビルド
make build

# リント
make lint
```
