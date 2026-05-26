# KRX Pension Bot

한국거래소(KRX)의 투자자별 매매 동향 중 `연기금등` 수급을 텔레그램 명령어로 조회하는 경량 Go 봇입니다.

이 프로젝트에서 말하는 `연기금등`은 국민연금 단독 매매가 아니라, KRX 투자자 분류상 `연기금등`으로 집계되는 수급 데이터입니다. 조회 결과는 투자 참고용이며 매매 추천이 아닙니다.

## 주요 기능

- 텔레그램 명령어 기반 수급 조회
- KOSPI / KOSDAQ 시장별 연기금등 순매수, 순매도 리포트
- 종목 코드별 연기금등 수급 조회
- Koyeb Web Service 배포를 전제로 한 webhook 방식
- 5~30분 메모리 캐시로 KRX 요청 부담 완화
- 텍스트 리포트 우선 제공

## 명령어

| 명령어 | 설명 |
| --- | --- |
| `/start` | 봇 소개 |
| `/help` | 사용 가능한 명령어 안내 |
| `/pension today` | 오늘 기준 시장별 연기금등 수급 |
| `/pension 5d` | 최근 5거래일 기준 시장별 연기금등 수급 |
| `/pension 20d` | 최근 20거래일 기준 시장별 연기금등 수급 |
| `/stock 005930` | 종목 코드 기준 오늘 수급 |
| `/stock 005930 20d` | 종목 코드 기준 최근 20거래일 수급 |

## 리포트 기준

- `/pension today`는 KOSPI와 KOSDAQ을 나누어 표시합니다.
- 기본 리포트는 각 시장별 순매수 TOP 10, 순매도 TOP 10입니다.
- 정렬 기준은 연기금등 순매수/순매도 금액입니다.
- `연기금등` 데이터는 국민연금 단독 매매 내역이 아닙니다.
- KRX 데이터 제공 상황, 장 마감 시각, 휴장일에 따라 조회 결과가 달라질 수 있습니다.

## 환경변수

| 이름 | 설명 | 예시 |
| --- | --- | --- |
| `TELEGRAM_BOT_TOKEN` | BotFather에서 발급받은 봇 토큰 | `123456:ABC...` |
| `TELEGRAM_WEBHOOK_SECRET` | Telegram webhook 요청 검증용 secret | `change-me` |
| `PUBLIC_BASE_URL` | Koyeb 서비스의 공개 URL | `https://krx-pension-bot.example.koyeb.app` |
| `CACHE_TTL_MINUTES` | KRX 응답 캐시 시간 | `10` |

`CACHE_TTL_MINUTES`는 무료 서버 운영을 고려해 5~30분 범위로 사용하는 것을 권장합니다.

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

### Koyeb

1. GitHub 저장소를 Koyeb Web Service로 연결합니다.
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
- 거래대금 대비 연기금등 순매수 비중 추가
- 종목별 기간 비교 리포트 강화
- 텔레그램 이미지 차트 리포트 추가
