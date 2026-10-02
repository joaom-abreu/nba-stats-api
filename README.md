# NBA Stats API

API de estatísticas de times e jogos da NBA, desenvolvida em Go.

## Escopo planejado

- Listagem de times e jogos com filtros e paginação.
- Resumo de desempenho por temporada e fase.
- Análise de jogos recentes e desempenho em casa e fora.
- Importação de dados com controle de progresso e retomada.
- Dados sintéticos para demonstração sem chave externa.

O recorte inicial de dados reais será a temporada iniciada em 2024,
com a fase explicitada nas importações e consultas.

## Estado atual

Servidor HTTP com endpoint de saúde e testes automatizados.

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

- [Regras de domínio e métricas](docs/domain-rules.md)