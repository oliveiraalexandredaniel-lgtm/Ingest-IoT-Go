# Ingest-O-IoT-Go: Ingestão Otimizada para IoT em Go

Middleware concorrente em Go para ingestão, pré-processamento em memória e persistência de telemetria industrial (IIoT).
Projeto de Iniciação Científica (FATEC SJC / LABSCA - ITA).

## Arquitetura

```
Daemons (sensores simulados) --MQTT--> Mosquitto --> Middleware Go --batch insert--> TimescaleDB
```

- **Daemons:** simuladores em Go que publicam telemetria via MQTT (`cmd/daemon`).
- **Broker:** Eclipse Mosquitto.
- **Middleware:** Goroutines/Channels, Worker Pool e janelamento temporal em memória (`internal/`, Etapas 2 e 3).
- **Banco:** PostgreSQL com a extensão TimescaleDB, persistência em lote.

## Estrutura

```
cmd/daemon/        simulador de sensores
internal/          middleware (etapa futura)
db/init/           SQL executado na 1ª subida do banco (extensão + hypertable)
mosquitto/config/  configuração do broker
docs/              relatório preliminar e levantamento bibliográfico
docker-compose.yml infraestrutura base
```

## Status

- [x] Etapa 1: ambiente base em Docker (broker + banco com esquema)
- [x] Simulador de sensores (protótipo)
- [ ] Etapa 1: relatório preliminar e levantamento bibliográfico (`docs/`)
- [ ] Etapa 2: middleware com Worker Pool e janelamento
- [ ] Etapa 3: integração com o banco e validação dos dados

## Como executar

Pré-requisitos: Docker com Compose e Go 1.22 ou superior.

```bash
cp .env.example .env          # ajuste as credenciais se quiser
docker compose up -d
docker compose ps             # os dois serviços devem aparecer como "healthy"
```

### Teste de fumaça

```bash
# Banco: extensão e hypertable criadas
docker exec timescale-db psql -U postgres -d iiot_db -c '\dx' -c '\d sensor_readings'

# Broker: assine em um terminal...
docker exec mqtt-broker mosquitto_sub -t 'iiot/#' -v
# ...e rode o simulador em outro
go mod tidy
go run ./cmd/daemon -sensors 5 -interval 500ms
```

Parâmetros do simulador: `-broker`, `-sensors`, `-interval`, `-qos`, `-topic-prefix`.

### Reiniciar do zero

Os scripts de `db/init/` só rodam com o volume vazio:

```bash
docker compose down -v
docker compose up -d
```

> A configuração do Mosquitto permite acesso anônimo e é só para desenvolvimento.
