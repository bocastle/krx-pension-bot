# KRX 주식 봇

한국거래소(KRX)의 투자자별 매매 동향을 텔레그램 명령어로 조회하는 경량 Go 봇입니다. 현재 MVP는 `연기금등` 수급 리포트를 중심으로 제공합니다.

이 프로젝트에서 말하는 `연기금등`은 국민연금 단독 매매가 아니라, KRX 투자자 분류상 `연기금등`으로 집계되는 수급 데이터입니다. 조회 결과는 투자 참고용이며 매매 추천이 아닙니다.

## 주요 기능

- 텔레그램 명령어 기반 수급 조회
- KOSPI / KOSDAQ 시장별 연기금등 순매수, 순매도 리포트
- 종목 코드, 종목명, 일부 별칭 기반 연기금등 수급 조회
- 거래대금 상위와 연기금등 순매수 상위를 묶은 관심 종목 리포트
- 거래대금 대비 연기금등 순매수 비중 표시
- 장전, 휴장일, KRX 데이터 지연 시 최근 조회 가능 거래일 자동 조회
- KRX 시간외 가격 기준 급등 종목 리포트
- Render Web Service 배포를 전제로 한 webhook 방식
- 5~30분 메모리 캐시로 KRX 요청 부담 완화
- 텍스트 리포트 우선 제공

## 명령어

| 명령어 | 설명 |
| --- | --- |
| `/시작` | 봇 소개 |
| `/도움말` | 사용 가능한 명령어 안내 |
| `/연기금 오늘` | 오늘 기준 시장별 연기금등 수급 |
| `/연기금 5일` | 최근 5거래일 기준 시장별 연기금등 수급 |
| `/연기금 10일` | 최근 10거래일 기준 시장별 연기금등 수급 |
| `/연기금 20일` | 최근 20거래일 기준 시장별 연기금등 수급 |
| `/연기금 오늘 20` | 오늘 기준 시장별 TOP 20 |
| `/관심 오늘` | 거래대금 상위와 연기금등 순매수 상위 요약 |
| `/거래대금 오늘` | 오늘 기준 시장별 거래대금 TOP 10 |
| `/수급상위 오늘` | 오늘 기준 시장별 연기금등 순매수 TOP 10 |
| `/시간외 급등` | KRX 시간외 가격 기준 시장별 급등 TOP 10 |
| `/시간외 급등 20` | KRX 시간외 가격 기준 시장별 급등 TOP 20 |
| `/종목 005930` | 종목 코드 기준 오늘 수급 |
| `/종목 삼성전자` | 종목명 기준 오늘 수급 |
| `/종목 삼전` | 자주 쓰는 별칭 기준 오늘 수급 |
| `/종목 하이닉스` | 일부 종목명 기준 오늘 수급 |
| `/종목 카뱅` | 별칭 기준 오늘 수급 |
| `/종목 두산에너` | 일부 종목명/별칭 기준 오늘 수급 |
| `/종목 005930 10일` | 종목 코드 기준 최근 10거래일 수급 |
| `/종목 005930 20일` | 종목 코드 기준 최근 20거래일 수급 |
| `삼성전자` | 종목명만 입력해 오늘 수급 조회 |

기존 영문 명령어(`/start`, `/help`, `/pension today`, `/stock 005930`, `/afterhours up`)도 함께 지원합니다.

## 리포트 기준

- `/pension today`는 KOSPI와 KOSDAQ을 나누어 표시합니다.
- 기본 리포트는 각 시장별 순매수 TOP 10, 순매도 TOP 10입니다.
- 정렬 기준은 연기금등 순매수/순매도 금액입니다.
- `/거래대금 오늘`은 KRX 시세 데이터의 거래대금 기준으로 정렬합니다.
- `/관심 오늘`은 외부 인기검색어가 아니라 거래대금 상위와 연기금등 순매수 상위를 함께 보여주는 요약 리포트입니다.
- 연기금등 수급 리포트와 관심 종목 리포트에는 가능한 경우 `거래대금 대비 연기금등 순매수 비중`을 함께 표시합니다.
- `/시간외 급등`은 네이버페이 증권 공개 시세의 KRX 시간외 가격 정보를 기준으로 표시합니다.
- 종목 조회는 종목코드, 정확한 종목명, 일부 별칭을 지원합니다. 여러 종목이 걸리는 검색어는 후보 목록을 보여줍니다.
- `연기금등` 데이터는 국민연금 단독 매매 내역이 아닙니다.
- KRX 데이터 제공 상황, 장 마감 시각, 휴장일에 따라 조회 결과가 달라질 수 있습니다. 오늘 데이터가 비어 있으면 최근 조회 가능한 거래일을 자동으로 찾아 표시합니다.

## 환경변수

| 이름 | 설명 | 예시 |
| --- | --- | --- |
| `TELEGRAM_BOT_TOKEN` | BotFather에서 발급받은 봇 토큰 | `123456:ABC...` |
| `TELEGRAM_WEBHOOK_SECRET` | Telegram webhook 요청 검증용 secret | `change-me` |
| `PUBLIC_BASE_URL` | Render 서비스의 공개 URL | `https://krx-pension-bot.onrender.com` |
| `CACHE_TTL_MINUTES` | KRX 응답 캐시 시간 | `10` |

`CACHE_TTL_MINUTES`는 무료 서버 운영을 고려해 5~30분 범위로 사용하는 것을 권장합니다.

## 로컬 실행

```bash
cp .env.example .env
```

`.env`에 Telegram 토큰과 webhook secret을 채운 뒤 실행합니다.

```bash
go run ./cmd/bot
```

서버가 뜨면 상태 확인 엔드포인트를 호출합니다.

```bash
curl http://localhost:8080/healthz
```

테스트는 다음 명령으로 실행합니다.

```bash
go test ./...
```

## 배포

### Telegram BotFather

1. Telegram에서 BotFather를 엽니다.
2. `/newbot`으로 봇을 생성합니다.
3. 발급받은 토큰을 `TELEGRAM_BOT_TOKEN`에 설정합니다.
4. 배포가 끝난 뒤 webhook URL을 등록합니다.

Webhook 등록 예시는 다음과 같습니다.

```bash
curl -X POST "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/setWebhook" \
  -d "url=<PUBLIC_BASE_URL>/telegram/webhook" \
  -d "secret_token=<TELEGRAM_WEBHOOK_SECRET>"
```

### Render

1. GitHub 저장소를 Render Web Service로 연결합니다.
2. Dockerfile 기반 배포를 선택합니다.
3. 필요한 환경변수를 설정합니다.
4. 배포 후 `/healthz`가 정상 응답하는지 확인합니다.
5. BotFather에서 발급받은 토큰으로 Telegram webhook을 등록합니다.

## 개발 방향

- Go 단일 HTTP 서버로 동작합니다.
- React나 별도 프론트엔드는 사용하지 않습니다.
- Telegram은 long polling 대신 webhook 방식을 사용합니다.
- KRX 요청 결과는 메모리에 TTL 캐시합니다.
- 차트 이미지는 이후 Go chart 라이브러리로 추가할 수 있도록 리포트 생성 구조를 분리합니다.

## 향후 개선

- `/pension today 20`처럼 표시 개수 확장
- 종목별 기간 비교 리포트 강화
- 텔레그램 이미지 차트 리포트 추가

## Render 무료 배포와 주기적 상태 확인

Render Free Web Service는 일정 시간 요청이 없으면 sleep 상태로 전환될 수 있습니다. 이 저장소에는 10분마다 `/healthz`를 호출하는 GitHub Actions 워크플로가 포함되어 있습니다.

Render 배포 후 생성된 서비스 URL을 GitHub 저장소 변수에 등록합니다.

```text
Settings > Secrets and variables > Actions > Variables > New repository variable
```

변수 이름은 다음과 같이 설정합니다.

```text
RENDER_HEALTH_URL
```

값은 Render 서비스의 health check URL로 설정합니다.

```text
https://<render-service-name>.onrender.com/healthz
```

워크플로는 `.github/workflows/render-health-ping.yml`에 있으며, `RENDER_HEALTH_URL`이 비어 있으면 아무 작업도 하지 않고 종료합니다. Render 배포가 끝난 뒤 이 변수를 설정하면 다음 스케줄부터 10분마다 health check 요청을 보냅니다.
