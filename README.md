# NBA Stats API

API de estatísticas de times e jogos da NBA, desenvolvida em Go.

## Escopo

- Listagem de times e jogos com filtros e paginação.
- Resumo de desempenho por temporada e fase.
- Análise de jogos recentes e desempenho em casa e fora.
- Importação de dados com controle de progresso e retomada.
- Dados sintéticos para demonstração sem chave externa.

O recorte inicial de dados reais será a temporada iniciada em 2024,
com a fase explicitada nas importações e consultas.

## Estado atual

Servidor HTTP com endpoint de saúde, validação de partidas e testes automatizados.
## Requisitos

Go na versão indicada em `go.mod` ou superior.

## Executar

Na raiz do projeto:

```bash
go run ./cmd/api
```

A API escuta na porta 8080.

Em outro terminal:

```bash
curl -i http://localhost:8080/health/live
```

Resposta esperada: HTTP 200 com o corpo:

```json
{"status":"ok"}
```

## Verificar

```bash
go test ./...
go vet ./...
```

## Documentação

- [Regras do projeto](docs/domain-rules.md)
- [Como a API vai funcionar](docs/api-contract.md)

## Banco de dados

Requer Docker com Docker Compose.

Na primeira execução, crie a configuração local:

```bash
cp .env.example .env
```

Para iniciar o PostgreSQL:

```bash
docker compose up -d --wait
```

O banco fica disponível em `127.0.0.1:5432`.
As configurações estão no `.env`.

Para abrir o terminal SQL:

```bash
docker compose exec db psql -U nba -d nba_stats
```

Use `\q` para sair.

Para parar os containers mantendo os dados:

```bash
docker compose down
```

## Migrações

Usamos Goose para acompanhar as mudanças do banco.

Instale a ferramenta na raiz do projeto:

```bash
mkdir -p bin
GOBIN="$PWD/bin" go install github.com/pressly/goose/v3/cmd/goose@v3.27.3
```

Com o banco iniciado e o `.env` configurado, aplique as migrações:

```bash
./bin/goose up
```

Para conferir quais foram aplicadas:

```bash
./bin/goose status
```

## Dados de demonstração

A demonstração usa quatro times fictícios e cinco jogos.

Com o banco iniciado e as migrações aplicadas, carregue os dados:

```bash
docker compose exec -T db psql -X -U nba -d nba_stats \
  -v ON_ERROR_STOP=1 < testdata/demo.sql
```

A carga pode ser repetida sem duplicar os registros.

Para conferir as quantidades e os resultados do time A:

```bash
docker compose exec -T db psql -X -U nba -d nba_stats \
  -v ON_ERROR_STOP=1 < testdata/check_demo.sql
```

São esperados quatro times, cinco jogos e, para o time A,
três jogos encerrados, duas vitórias e uma derrota.