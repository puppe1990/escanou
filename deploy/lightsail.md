# Deploy no AWS Lightsail (HTTPS + câmera)

O escaneamento de código de barras no celular exige **HTTPS** e permissão de câmera. Este guia usa Lightsail com Caddy como proxy TLS.

## 1. Instância Lightsail

1. Crie uma instância **Ubuntu 22.04** (plano mínimo serve para MVP).
2. Anexe um **IP estático**.
3. Abra as portas **80** e **443** no firewall da instância.
4. Aponte o DNS do domínio (`A`) para o IP estático.

## 2. Build e envio da imagem

Na máquina de build (com Docker):

```bash
cd /caminho/para/escanou
docker build -t escanou:latest .
docker save escanou:latest | gzip > escanou.tar.gz
scp escanou.tar.gz ubuntu@SEU_IP:/tmp/
```

Na instância:

```bash
docker load < /tmp/escanou.tar.gz
mkdir -p /opt/escanou/data
```

## 3. Variáveis de produção

Crie `/opt/escanou/.env`:

```bash
ENV=production
PORT=:8080
APP_URL=https://seu-dominio.com
DB_PATH=/app/data/app.db
ADMIN_TOKEN=<gere-um-token-forte>
LOCALE=pt

# Câmera no PWA (obrigatório para escanear no celular)
PERMISSIONS_POLICY=camera=(self)
CSP_MEDIA_SRC=blob:
CSP_CONNECT_SRC=https://world.openfoodfacts.org

# Se usar proxy reverso na mesma máquina
TRUSTED_PROXIES=127.0.0.1
```

## 4. Caddy (TLS automático)

```bash
sudo apt update && sudo apt install -y caddy
```

`/etc/caddy/Caddyfile`:

```
seu-dominio.com {
  reverse_proxy 127.0.0.1:8080
}
```

```bash
sudo systemctl reload caddy
```

## 5. Container da aplicação

`/opt/escanou/docker-compose.yml`:

```yaml
services:
  escanou:
    image: escanou:latest
    restart: unless-stopped
    env_file: .env
    volumes:
      - ./data:/app/data
    ports:
      - "127.0.0.1:8080:8080"
```

```bash
cd /opt/escanou
docker compose up -d
```

## 6. Verificação

```bash
curl -I https://seu-dominio.com/
curl -I https://seu-dominio.com/ | grep -i permissions-policy
# deve conter camera=(self)

curl -I https://seu-dominio.com/ | grep -i content-security-policy
# deve conter media-src 'self' blob:
```

No celular:

1. Abra `https://seu-dominio.com/login` e entre.
2. Vá em **Escanear Preço** — o navegador pedirá permissão de câmera.
3. Adicione à tela inicial (PWA `standalone`).

## 7. Worker de jobs (opcional)

Rode um segundo serviço no compose com o mesmo volume de dados e `cais jobs work` (ver [jobs design](https://github.com/puppe1990/cais)).

## Troubleshooting

| Sintoma                           | Correção                                          |
| --------------------------------- | ------------------------------------------------- |
| Câmera bloqueada                  | `PERMISSIONS_POLICY=camera=(self)` + HTTPS        |
| Scanner não inicia                | `CSP_MEDIA_SRC=blob:` no `.env`                   |
| Produto desconhecido não cadastra | `CSP_CONNECT_SRC=https://world.openfoodfacts.org` |
| IP errado nos logs/rate limit     | `TRUSTED_PROXIES` com IP do proxy                 |
