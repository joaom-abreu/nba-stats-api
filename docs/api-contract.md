# Como a API vai funcionar

Estas são as rotas planejadas. A documentação será atualizada
conforme elas forem implementadas.

## Rotas

| Rota GET | O que faz |
|---|---|
| `/v1/teams` | Lista times |
| `/v1/teams/{id}` | Mostra um time |
| `/v1/games` | Lista jogos |
| `/v1/teams/{id}/summary` | Mostra o desempenho do time |
| `/v1/teams/{id}/recent` | Mostra os últimos jogos e suas médias |
| `/v1/import-runs/{id}` | Mostra o andamento de uma importação |
| `/health/live` | Confere se a aplicação responde |
| `/health/ready` | Confere o banco e o esquema |

Os IDs usados nas rotas são IDs locais, inteiros positivos.

## Filtros

Nas listagens e análises, informe `source`:

- `demo` para dados fictícios.
- `balldontlie` para dados reais.

Times podem ser filtrados por conferência: `east` ou `west`.

Jogos podem ser filtrados por:

- `team_id`: mandante ou visitante.
- `season`: ano inicial da temporada.
- `phase`: inicialmente `regular_season` ou `playoffs`.
- `status`: estado da partida.
- `date_from` e `date_to`: datas inclusivas no formato YYYY-MM-DD.

No resumo e nos últimos jogos, origem, temporada e fase
são obrigatórias. As datas são opcionais.

Exemplo:

`/v1/games?source=demo&team_id=1&season=2024&phase=regular_season`

## Paginação

As listas de times e jogos usam:

- `limit`: padrão 20, entre 1 e 100.
- `offset`: padrão 0, sem valores negativos.

Times são ordenados por ID crescente.
Jogos são ordenados por data decrescente e ID decrescente.

Últimos jogos usa apenas `limit=5` ou `limit=10`, com padrão 5.

## Respostas

As listas têm:

- `data`: registros encontrados.
- `meta`: limit, offset e quantidade retornada (`returned`).

Uma lista vazia retorna HTTP 200 com `data: []`.

O resumo contém `team_id` e os grupos `overall`, `home` e `away`.
Cada grupo informa:

- `games`, `wins` e `losses`.
- `win_rate`.
- `avg_points_for` e `avg_points_against`.
- `avg_point_difference`.

Últimos jogos retorna `team_id`, a lista `games` e suas `metrics`.

As respostas analíticas também informam o recorte usado,
a última coleta disponível e a quantidade de jogos considerada.

## Erros

- 400: parâmetro ausente ou inválido.
- 404: recurso solicitado por ID não encontrado.
- 405: método não permitido.
- 500: erro inesperado.
- 503: banco temporariamente indisponível.

Os erros dos endpoints de negócio terão código, mensagem
e identificador da requisição para ajudar a localizar os logs.

## Regras dos cálculos

Veja [Regras do projeto](domain-rules.md).