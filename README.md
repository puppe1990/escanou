# Escanou

Comparador de preços de supermercado — app [Cais](https://github.com/puppe1990/cais) com escaneamento de código de barras, feed colaborativo e gamificação.

## Funcionalidades

| Rota            | Descrição                                        |
| --------------- | ------------------------------------------------ |
| `/`             | Escanear preço (câmera + lookup Open Food Facts) |
| `/map`          | Mapa de supermercados                            |
| `/feed`         | Feed de preços reportados pela comunidade        |
| `/achievements` | Conquistas e ranking                             |
| `/nfce`         | Importar nota fiscal (NFC-e)                     |
| `/login`        | Autenticação (sessão, 7 dias)                    |

Demo em desenvolvimento: `demo@example.com` / `password`

## Stack

- Go 1.26 + Cais (HTMX, Tailwind, SQLite)
- PWA com escaneamento via `html5-qrcode`
- Open Food Facts para lookup de produtos

## Setup

Depende de [github.com/puppe1990/cais](https://github.com/puppe1990/cais) v0.6.0+.

Para desenvolver com o framework local:

```bash
go work init .
go work use ../../Cais   # ajuste o path para o clone do Cais
```

```bash
cd escanou
cp .env.example .env
cais install
cais dev          # http://localhost:8080
cais test
```

## Variáveis de ambiente

| Variável             | Default         | Descrição                            |
| -------------------- | --------------- | ------------------------------------ |
| `PORT`               | `:8080`         | Porta do servidor                    |
| `DB_PATH`            | `./data/app.db` | SQLite                               |
| `ENV`                | `development`   | `development` ou `production`        |
| `APP_URL`            | —               | Obrigatório em produção (OG/PWA)     |
| `LOCALE`             | `en`            | `pt` ou `en`                         |
| `PERMISSIONS_POLICY` | —               | `camera=(self)` para scan no celular |
| `CSP_MEDIA_SRC`      | —               | `blob:` para câmera no PWA           |
| `CSP_CONNECT_SRC`    | —               | `https://world.openfoodfacts.org`    |

Ver `.env.example` para produção (`ADMIN_TOKEN`, SMTP, etc.).

## Deploy

Guia AWS Lightsail com HTTPS e câmera: [`deploy/lightsail.md`](deploy/lightsail.md)

```bash
docker build -t escanou:latest .
```

## Estrutura

```
internal/app/       → rotas e bootstrap
internal/handlers/  → supermarket (scan, feed, map…), auth
internal/store/     → SQLite + domínio de preços
web/templates/      → HTML + partials (icons, nav, scan)
web/static/js/      → scan.js, html5-qrcode
```

## CI

GitHub Actions: `go test`, `golangci-lint`, Prettier.

```bash
make pre-commit-install
make ci
```
