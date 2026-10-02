# NBA Stats API

Projeto de estudo para praticar Go, SQL e Docker criando uma API de estatísticas da NBA.

## Tecnologias

- Go
- PostgreSQL
- Docker Compose
- pgxpool para conexão com o banco
- Goose para migrações

## Requisitos

- Go 1.25 ou superior.
- Docker com Docker Compose disponível.

Execute os comandos abaixo na raiz do projeto.

## Configuração

Na primeira execução, copie o arquivo de exemplo:

```bash
cp .env.example .env
```

Se o `.env` já existir, use o arquivo atual e confira se ele contém `DATABASE_URL`.

A API lê o `.env` ao iniciar. As configurações principais são:

- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` e `POSTGRES_PORT`: configuração do banco no Docker.
- `DATABASE_URL`: conexão usada pela aplicação Go.
- `GOOSE_DBSTRING`: conexão usada pelas migrações.

Se alterar usuário, senha, banco ou porta, ajuste também as duas URLs de conexão.

Baixe as dependências do projeto:

```bash
go mod download
```

## Banco de dados

Suba o PostgreSQL:

```bash
docker compose up -d --wait
```

Instale o Goose na pasta `bin`:

```bash
mkdir -p bin
GOBIN="$PWD/bin" go install github.com/pressly/goose/v3/cmd/goose@v3.27.3
```

Aplique as migrações e confira o status:

```bash
./bin/goose up
./bin/goose status
```

Para abrir o terminal do PostgreSQL com a configuração padrão:

```bash
docker compose exec db psql -X -U nba -d nba_stats
```

Use `\q` para sair. Se personalizou o usuário ou o nome do banco, ajuste os comandos `psql` deste README.

## Dados de demonstração

Os dados são fictícios e não precisam de chave de API. Eles usam a origem `demo`, a temporada `2024` e a fase `regular_season`.

Carregue os dados:

```bash
docker compose exec -T db psql -X -U nba -d nba_stats \
  -v ON_ERROR_STOP=1 < testdata/demo.sql
```

O comando pode ser repetido: ele atualiza os registros existentes, sem duplicar times ou partidas.

Confira os resultados:

```bash
docker compose exec -T db psql -X -U nba -d nba_stats \
  -v ON_ERROR_STOP=1 < testdata/check_demo.sql
```

O resultado esperado é de **4 times e 5 partidas**. Para o time A, considerando somente partidas finalizadas:

| Estatística | Resultado |
| --- | --- |
| Partidas | 3 |
| Vitórias | 2 |
| Derrotas | 1 |
| Média de pontos feitos | 100,00 |
| Média de pontos sofridos | 100,00 |

## Executar a API

Com o banco ligado e o `.env` configurado:

```bash
go run ./cmd/api
```

A aplicação verifica a conexão com o PostgreSQL antes de iniciar o servidor. Se a conexão falhar, ela mostra um erro e encerra.

Em outro terminal, consulte:

```bash
curl -i http://localhost:8080/health/live
```

A resposta deve ter status `200` e este corpo:

```json
{"status":"ok"}
```

Esse endpoint indica que o servidor está respondendo. A conexão com o banco é verificada na inicialização; a verificação pelo endpoint `/health/ready` será implementada depois.

Para encerrar a API, pressione `Ctrl+C`.

## Testes

Execute os testes e a análise do código:

```bash
go test ./...
go vet ./...
```

Para testar a conexão com o PostgreSQL, deixe o banco ligado e execute:

```bash
go test -tags=integration ./internal/postgres
```

O teste usa `DATABASE_URL` do `.env`. Também é possível definir `TEST_DATABASE_URL` no ambiente para escolher outro banco.

## Parar o banco

```bash
docker compose down
```

Os dados permanecem no volume do Docker e estarão disponíveis quando o banco subir novamente.

## Documentação

- [Regras do projeto](docs/domain-rules.md)
- [Como a API vai funcionar](docs/api-contract.md)