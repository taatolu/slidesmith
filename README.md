# slidesmith

プロンプトから PowerPoint 資料を生成し、ユーザー本人の Google ドライブへ保存する AI エージェントアプリです。

ユーザーから「Google アカウントの OAuth トークン付きプロンプト」を受け取り、AI エージェントに PPTX 生成用の Python スクリプトを書かせ、それを Cloud Run 上の Python 実行環境で実行して、できあがった `.pptx` をユーザーのマイドライブに返却します。

`sampleAgentApp`（Google Workspace から Drive / Gemini Enterprise へのプロキシ）と同じ設計方針（クリーンアーキテクチャ、最小権限、ユーザートークン移譲）を引き継いでいます。

---

## 1. 全体像

```
[ フロントエンド (GAS / Web) ]
        │  ① プロンプト + OAuth アクセストークン
        ▼
┌──────────────────────────────┐
│ Cloud Run ① api (Go)         │
│  ・トークン検証（introspect） │
│  ・エージェント呼び出し       │
│  ・Drive へアップロード       │
└──────────────────────────────┘
   │ ②生成依頼            ▲ ⑤ PPTX バイト列
   ▼                      │
[ AI エージェント ]   ┌──────────────────────────────┐
  (Gemini Enterprise) │ Cloud Run ② executor (Python)│
   │ ③ Python コード   │  ・python-pptx で実行        │
   └──────────────────▶│  ・外部通信なし／認証情報なし │
      ④ 実行依頼       └──────────────────────────────┘
                                   │
                        ⑥ マイドライブへ保存（ユーザー権限）
                                   ▼
                          [ ユーザーの Google Drive ]
```

### 処理の流れ

| # | 処理 | 担当 |
|---|---|---|
| ① | プロンプトとアクセストークンを `Authorization` ヘッダで受信 | api |
| ② | トークンを検証し、必要なスコープ（`drive.file` 等）を確認 | api |
| ③ | エージェントにスライド構成を渡し、`python-pptx` のコードを生成 | api → AI エージェント |
| ④ | 生成コードを executor に POST し、サンドボックス内で実行 | api → executor |
| ⑤ | 実行結果の `.pptx` をバイト列（Base64）で受け取る | executor → api |
| ⑥ | ユーザーのアクセストークンで Drive API を呼び、マイドライブへ保存 | api |
| ⑦ | 保存したファイルの `fileId` と `webViewLink` を返却 | api |

---

## 2. サービスを 2 つに分ける理由

AI が生成したコードを実行する以上、そのコードは「信用できないコード」として扱います。

| | api (Go) | executor (Python) |
|---|---|---|
| 役割 | 認証・オーケストレーション・Drive 連携 | 生成コードの実行のみ |
| ユーザートークン | 保持する | **渡さない** |
| 外部ネットワーク | 許可（Vertex AI / Drive API） | **遮断**（VPC egress 制限） |
| 実行時 SA | `slidesmith-api-sa` | `slidesmith-exec-sa`（権限なし） |
| 呼び出し元 | フロントエンド（要認証） | api からのみ（IAM 認証必須・`--no-allow-unauthenticated`） |

executor に認証情報もネットワークも与えないことで、生成コードが暴走してもユーザーのドライブや他リソースには到達できません。

---

## 3. ディレクトリ構成

```
slidesmith/
├── api/                      // Cloud Run ①: Go / クリーンアーキテクチャ
│   ├── cmd/api/main.go
│   ├── domain/
│   │   ├── model/            // oauth_token.go, deck.go, script.go, error.go
│   │   ├── service/          // auth_service.go
│   │   └── repository/       // drive_repository.go, agent_repository.go, executor_repository.go
│   ├── usecase/              // generate_deck_usecase.go
│   ├── infrastructure/
│   │   ├── gws/              // Google Drive API
│   │   ├── agent/            // Gemini Enterprise / Vertex AI
│   │   └── executor/         // executor サービスへの HTTP クライアント
│   ├── presentation/http/    // handler / router / request / response
│   ├── registry/             // DI
│   ├── config/
│   ├── mocks/
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── executor/                 // Cloud Run ②: Python 実行環境
│   ├── main.py               // FastAPI: POST /execute
│   ├── sandbox.py            // タイムアウト・メモリ制限・一時ディレクトリ
│   ├── requirements.txt      // python-pptx, fastapi, uvicorn
│   └── Dockerfile
├── infra/                    // Terraform / デプロイスクリプト
├── docs/                     // 設計メモ、プロンプト設計
└── README.md
```

---

## 4. API 仕様（予定）

### `POST /v1/decks`

リクエスト

```http
POST /v1/decks HTTP/1.1
Authorization: Bearer <ユーザーの OAuth アクセストークン>
Content-Type: application/json

{
  "prompt": "2026年度の事業計画を10枚のスライドにまとめて",
  "filename": "2026年度事業計画.pptx",
  "folderId": null
}
```

レスポンス

```json
{
  "fileId": "1AbC...",
  "fileName": "2026年度事業計画.pptx",
  "webViewLink": "https://docs.google.com/presentation/d/1AbC.../edit",
  "slideCount": 10
}
```

必要な OAuth スコープ

| スコープ | 用途 |
|---|---|
| `https://www.googleapis.com/auth/drive.file` | 本アプリが作成したファイルの作成・更新（最小権限） |
| `openid` / `email` | 実行主体の識別 |

### `POST /execute`（executor・内部用）

api からのみ呼び出します。生成された Python コードを受け取り、`output.pptx` を Base64 で返します。この受け渡し方式の理由と、将来の代替案は「9. PPTX の受け渡し方式」を参照してください。

---

## 5. 環境変数

### api

| 変数 | 説明 |
|---|---|
| `PROJECT_ID` | GCP プロジェクト ID |
| `REGION` | `asia-northeast1` |
| `AGENT_ENDPOINT` | Gemini Enterprise / Vertex AI のエンドポイント |
| `EXECUTOR_URL` | executor サービスの URL |
| `MAX_SLIDES` | 1 リクエストあたりの最大スライド数 |

### executor

| 変数 | 説明 |
|---|---|
| `EXEC_TIMEOUT_SEC` | スクリプト実行のタイムアウト（既定 60 秒） |
| `MAX_OUTPUT_BYTES` | 返却する PPTX の上限サイズ |

---

## 6. デプロイ

executor を先にデプロイし、その URL を api に渡します。

```bash
# ② executor（内部専用・未認証アクセス禁止）
gcloud run deploy slidesmith-executor \
  --source ./executor \
  --service-account=slidesmith-exec-sa@<PROJECT_ID>.iam.gserviceaccount.com \
  --region=asia-northeast1 \
  --no-allow-unauthenticated \
  --set-env-vars=EXEC_TIMEOUT_SEC=60

# ① api
gcloud run deploy slidesmith-api \
  --source ./api \
  --build-service-account=projects/<PROJECT_ID>/serviceAccounts/cloud-build-worker-sa@<PROJECT_ID>.iam.gserviceaccount.com \
  --service-account=slidesmith-api-sa@<PROJECT_ID>.iam.gserviceaccount.com \
  --region=asia-northeast1 \
  --set-env-vars=EXECUTOR_URL=<executor の URL>
```

api の SA に、executor を呼び出すための `roles/run.invoker` を付与します。

```bash
gcloud run services add-iam-policy-binding slidesmith-executor \
  --member=serviceAccount:slidesmith-api-sa@<PROJECT_ID>.iam.gserviceaccount.com \
  --role=roles/run.invoker \
  --region=asia-northeast1
```

---

## 7. 実装の進め方

- [ ] `api`: ドメイン層（`OauthToken`, `Deck`, `GeneratedScript`, `DomainError`）
- [ ] `api`: トークン検証（OAuth2 introspect / tokeninfo）の実装
- [ ] `executor`: `POST /execute` とサンドボックス（タイムアウト・一時ディレクトリ）
- [ ] `api`: エージェント連携とプロンプト設計（生成コードの制約をプロンプトで明示）
- [ ] `api`: Drive アップロード（ユーザートークンでの `files.create`）
- [ ] エンドツーエンド疎通、`infra/` の Terraform 化

---

## 8. 設計上の注意点

- **生成コードを api 側で実行しない。** executor 経由のみとします。
- **executor にユーザートークンを渡さない。** Drive への書き込みは必ず api が行います。
- **Cloud Run のファイルシステムは揮発性（インメモリ）です。** 生成した PPTX は一時ディレクトリに置き、Drive へ渡したら破棄します。サイズはメモリ上限に直結するため `MAX_OUTPUT_BYTES` で制限します。
- **`drive.file` スコープを基本とします。** `drive` フルスコープはユーザーの全ファイルへのアクセスを意味するため、避けます。
- **スコープがある＝アクセスできる、ではありません。** 実際の可否は Drive API 呼び出しの結果で判断します。

---

## 9. PPTX の受け渡し方式（executor → api）

### 9.1 現時点の方式：JSON + Base64

executor が生成した `.pptx` は、JSON レスポンスに Base64 文字列として載せて api に返します。

Base64 はバイナリを文字列に変換する都合上、**データサイズが約 1.33 倍に膨らみます**。ただし、通常のスライド生成（テキスト中心、画像が数枚）であれば PPTX は数 MB 程度に収まるため、この膨張は実用上まったく問題になりません。

実装のシンプルさと堅牢性を優先し、現時点ではこの方式を採用します。

- JSON 1 本でやり取りが完結し、エラー情報（実行ログ、スタックトレース）を同じレスポンスに同梱できる
- マルチパートやストリーミングの取り回しが不要で、Go / Python 双方で標準ライブラリだけで扱える

### 9.2 今後「重い」と感じた場合の代替策

将来的に「高画質な画像を大量に埋め込んだ 50MB〜100MB 超のスライド」を生成するようになった場合は、以下の設計変更を検討します。

| 代替策 | 内容 | トレードオフ |
|---|---|---|
| バイナリストリーミング | Base64 に変換せず、`application/octet-stream` で生のバイト列をそのままレスポンスボディとして返す | データ膨張なし。ただしエラー情報を本文に載せられないため、ヘッダやステータスコードでの表現が必要 |
| Cloud Storage 経由 | executor が生成ファイルを GCS にアップロードし、api には URL（または署名付き URL）だけを返す | メモリを介さず大容量に対応できる。ただし **executor に GCS への書き込み権限を与える必要があり、「executor に権限を持たせない」というサンドボックス方針とのトレードオフ**になる |

### 9.3 判断の目安

切り替えを検討する目安は次のとおりです。

- 生成される PPTX が恒常的に数十 MB を超える
- Cloud Run のメモリ上限（api / executor 双方でバイト列を保持する分）が逼迫する
- レスポンスサイズが Cloud Run の上限（HTTP/1 で 32MB）に近づく

いずれにも当てはまらないうちは、9.1 の Base64 方式を維持します。
