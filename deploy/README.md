# 幹事ちゃん (Kanji-Chan) デプロイガイド (GHCR Docker イメージ)

GitHub Container Registry (`ghcr.io/sweetfish329/kanji-chan:latest`) で公開されているビルド済みコンテナイメージを使用して、本番・検証環境へデプロイする手順です。

本構成では、Go バックエンドと Svelte 5 フロントエンドが1つの軽量コンテナにまとめられており、SQLite データベースをボリューム永続化して動作します。

---

## 1. 事前準備

デプロイ先サーバーに以下がインストールされていることを確認してください。

- **Docker** (20.10+)
- **Docker Compose** v2 (`docker compose` コマンド)

また、Google または GitHub で **OAuth 2.0 アプリケーション** を作成し、クライアントIDとクライアントシークレットを取得してください。
- **コールバックURL例**: `https://kanji.example.com/api/auth/callback`

---

## 2. ディレクトリ準備と環境変数設定

サーバー上の任意の作業ディレクトリに `deploy/` 配下のファイルを用意します。

```bash
mkdir -p kanji-chan && cd kanji-chan

# compose.yaml と .env.example をダウンロード（または手動配置）
curl -sSL -O https://raw.githubusercontent.com/sweetfish329/kanji-chan/main/deploy/compose.yaml
curl -sSL -O https://raw.githubusercontent.com/sweetfish329/kanji-chan/main/deploy/.env.example

# .env ファイルを作成
cp .env.example .env
```

`.env` ファイルを開き、環境に合わせて編集します。

```env
PORT=8080
DB_TYPE=sqlite
DB_PATH=/data/kanji.db

OAUTH_PROVIDER=google
OAUTH_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
OAUTH_CLIENT_SECRET=your-google-client-secret
OAUTH_REDIRECT_URI=https://kanji.example.com/api/auth/callback

# セッションキー (openssl rand -base64 32 などで生成)
SESSION_SECRET=super-secret-random-key-string

PUBLIC_SITE_URL=https://kanji.example.com
```

---

## 3. コンテナの起動

以下のコマンドを実行して最新イメージを取得し、バックグラウンドで起動します。

```bash
# 最新イメージの取得
docker compose pull

# コンテナ起動
docker compose up -d

# 起動状態の確認
docker compose ps

# ログ確認
docker compose logs -f
```

---

## 4. コンテナの更新 (アップデート手順)

GitHub Actions の CI/CD により `main` ブランチに更新が入ると、新しい Docker イメージが `ghcr.io` に自動プッシュされます。
サーバー側のイメージ更新と再起動は以下のコマンドで完了します。

```bash
# 最新イメージの取得とコンテナの再起動（ローリングアップデート）
docker compose pull && docker compose up -d
```

不要になった旧イメージを削除する場合は以下を実行します：

```bash
docker image prune -f
```

---

## 5. リバースプロキシ設定例 (HTTPS / SSL対応)

本番運用で Nginx や Caddy、Traefik などをリバースプロキシとして前段に配置する場合の設定例です。

### Nginx 設定例 (`/etc/nginx/sites-available/kanji.conf`)

```nginx
server {
    listen 80;
    server_name kanji.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name kanji.example.com;

    ssl_certificate /etc/letsencrypt/live/kanji.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/kanji.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy 設定例 (`Caddyfile`)

```caddyfile
kanji.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

---

## 6. バックアップ & リストア

データベース (`kanji.db`) は Docker Named Volume `kanji_data` (`/data`) に保存されます。

### バックアップ

```bash
# コンテナを一時停止して安全にSQLiteファイルをコピー
docker compose stop app
docker run --rm -v kanji-chan_kanji_data:/data -v $(pwd):/backup ubuntu tar cvzf /backup/kanji_db_backup.tar.gz /data
docker compose start app
```

### リストア

```bash
docker compose down
docker run --rm -v kanji-chan_kanji_data:/data -v $(pwd):/backup ubuntu tar xvzf /backup/kanji_db_backup.tar.gz -C /
docker compose up -d
```
