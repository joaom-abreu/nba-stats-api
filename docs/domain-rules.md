# Regras de domínio e métricas

Este documento define o comportamento esperado da NBA Stats API.

## 1. Temporada, fase e origem

- A temporada é identificada pelo ano em que começa.
- A temporada `2024` pode conter partidas realizadas em 2025.
- A fase faz parte do recort# Regras do projeto

## Dados

- Vamos começar pela temporada 2024, que também inclui jogos de 2025.
- A temporada e a fase precisam estar identificadas nos dados.
- `demo` identifica os dados fictícios.
- `balldontlie` identifica os dados reais.
- Cada registro é identificado pela origem e pelo ID externo.
- Importar o mesmo jogo novamente atualiza o registro existente.

## Jogos

Os estados possíveis são:

- `scheduled`: agendado.
- `in_progress`: em andamento.
- `finished`: encerrado.
- `postponed`: adiado.
- `cancelled`: cancelado.

Mandante e visitante precisam existir, ser diferentes e pertencer
à mesma origem do jogo.

Placares não podem ser negativos. Um placar ausente é `null`;
zero é um placar válido.

Um jogo encerrado precisa ter os dois placares e um vencedor.
Estados desconhecidos e resultados inválidos devem ser registrados
como rejeições na importação.

## Estatísticas

Só jogos encerrados e válidos entram nas estatísticas.

- Aproveitamento = vitórias / jogos.
- Média de pontos = soma dos pontos / jogos.
- Saldo médio = média de pontos feitos - média de pontos sofridos.
- Aproveitamento é uma fração: `0.6` significa 60%.
- Casa e fora são calculados separadamente.

Sem jogos, os contadores são zero e as médias e o aproveitamento
são `null`.

As consultas informam a origem, a temporada, a fase, as datas,
a quantidade de jogos e a última coleta disponível.
Os resultados representam os dados importados.

Os filtros de data incluem os dois limites.
Horários, quando disponíveis, são retornados em UTC.

## Últimos jogos

Selecionamos até 5 ou 10 jogos encerrados dentro do recorte.
O padrão é 5.

A ordem é da data mais recente para a mais antiga.
Em datas iguais, usamos o ID local decrescente.

Se houver apenas 3 jogos, as médias usam esses 3.

## Exemplo para conferir os cálculos

Dados fictícios do time A, na temporada 2024 e fase regular.
Os placares abaixo estão do ponto de vista de A.
Os números 101 a 105 são IDs externos da demonstração.
Os IDs locais são gerados pelo banco.

| ID  externo| Data | Local | Placar de A | Estado |
|---|---|---|---|---|
| 101 | 2025-01-02 | Casa | 110–100 | Encerrado |
| 102 | 2025-01-04 | Fora | 90–105 | Encerrado |
| 103 | 2025-01-06 | Casa | 100–95 | Encerrado |
| 104 | 2025-01-08 | Casa | Ausente | Agendado |
| 105 | 2025-01-10 | Fora | 62–60 | Em andamento |

Resultado esperado:

- 3 jogos, 2 vitórias e 1 derrota.
- Aproveitamento: 2/3.
- Média de pontos feitos: 100.
- Média de pontos sofridos: 100.
- Em casa: médias de 105 feitos e 97.5 sofridos.
- Fora: médias de 90 feitos e 105 sofridos.
- Últimos jogos: 103, 102 e 101.

Entre 4 e 6 de janeiro, entram os jogos 102 e 103:
1 vitória, 1 derrota, média de 95 feitos e 100 sofridos.

Entre 8 e 10 de janeiro, não há jogos encerrados:
contadores zero e médias e aproveitamento `null`.e e não é deduzida pela data da partida.
- O primeiro recorte real planejado é a temporada `2024`,
  na fase `regular_season`.
- A classificação da fase deve ser validada no adaptador da fonte.
  O campo externo `postseason`, sozinho, não define todas as fases.
- Dados sintéticos usam a origem `demo`.
- Dados reais da BALLDONTLIE usam a origem `balldontlie`.
- Consultas analíticas selecionam uma origem explícita para evitar
  misturar dados sintéticos e reais.

## 2. Identidade dos registros

- Times e jogos possuem IDs locais usados pela nossa API.
- Os IDs externos são preservados para identificar registros da fonte.
- A combinação de origem e ID externo é única em cada tipo de registro.
- Reimportar um jogo atualiza o registro existente.
- Uma correção de placar deve aparecer nas análises seguintes.

## 3. Estados das partidas

Os valores internos propostos são:

| Estado | Significado | Entra nas métricas? |
|---|---|---|
| `scheduled` | Partida agendada | Não |
| `in_progress` | Partida em andamento | Não |
| `finished` | Partida encerrada | Sim, com placares válidos |
| `postponed` | Partida adiada | Não |
| `cancelled` | Partida cancelada | Não |

O adaptador externo traduz os estados da fonte para esses valores.
Um estado desconhecido deve gerar uma rejeição identificável,
sem ser convertido silenciosamente em partida encerrada.

## 4. Validade de uma partida

- Mandante e visitante devem existir e ser diferentes.
- Os times devem pertencer à mesma origem da partida.
- Placares informados devem ser inteiros não negativos.
- Um placar ausente é `null`; zero é um valor válido.
- Partidas encerradas devem ter os dois placares.
- Uma partida encerrada com placares iguais é inconsistente
  e deve ser rejeitada, sem gerar empate nas análises.
- Uma partida em andamento pode ter placares, mas continua
  fora das métricas de resultados.

## 5. Recorte das consultas

O recorte analítico inclui:

- Origem.
- Temporada.
- Fase.
- Time.
- Data inicial e final, quando informadas.

As datas inicial e final são inclusivas e usam a data da partida
fornecida pela origem.

Horários de início, quando disponíveis, são armazenados com fuso
e serializados em UTC.

Os resultados representam os jogos disponíveis no banco local.
Não representam automaticamente a classificação oficial completa.

As respostas analíticas devem informar o recorte, a quantidade
de jogos considerados e a data da última coleta disponível.
Essa data não garante que todos os jogos do recorte foram importados.

## 6. Métricas por time

Somente partidas encerradas e válidas entram nos cálculos.

Para cada partida, os placares são interpretados do ponto de vista
do time consultado:

- Pontos a favor: pontos marcados pelo time.
- Pontos contra: pontos marcados pelo adversário.
- Vitória: pontos a favor maiores que pontos contra.
- Derrota: pontos a favor menores que pontos contra.

As métricas são:

- Jogos: quantidade de partidas encerradas válidas.
- Vitórias: quantidade de vitórias.
- Derrotas: quantidade de derrotas.
- Aproveitamento: vitórias divididas por jogos.
- Média de pontos a favor: soma dos pontos a favor dividida por jogos.
- Média de pontos contra: soma dos pontos contra dividida por jogos.
- Saldo médio: média de pontos a favor menos média de pontos contra.

O aproveitamento será representado na API como uma fração entre
0 e 1. Por exemplo, `0.6` corresponde a 60%.

Os cálculos usam divisão decimal, sem truncamento inteiro.
Arredondamentos de apresentação não devem alterar os cálculos.

O resumo em casa considera apenas jogos como mandante.
O resumo fora considera apenas jogos como visitante.

## 7. Time sem jogos encerrados

Quando não houver partidas encerradas válidas no recorte:

- Jogos, vitórias e derrotas são zero.
- Aproveitamento é `null`.
- Médias de pontos e saldo médio são `null`.

A ausência de resultados não deve ser apresentada como média zero.

Essa regra também vale separadamente para os grupos casa e fora.

## 8. Últimos jogos

- Aplicar primeiro o recorte solicitado.
- Considerar somente partidas encerradas válidas.
- Ordenar pela data da partida em ordem decrescente.
- Em datas iguais, desempatar pelo ID local em ordem decrescente.
- Aceitar `limit=5` ou `limit=10`, com padrão igual a 5.
- Aplicar o limite antes de calcular as médias.
- Informar quantos jogos foram efetivamente encontrados.

Se houver apenas três jogos e o limite for cinco, a resposta
terá três jogos e as médias serão calculadas sobre esses três.

## 9. Exemplo sintético para conferência

Todos os jogos abaixo pertencem à origem `demo`,
à temporada `2024` e à fase `regular_season`.

Os times A, B, C e D são fictícios.

| ID local | Data | Mandante | Visitante | Pontos mandante | Pontos visitante | Estado |
|---|---|---|---|---|---|---|
| 101 | 2025-01-02 | A | B | 110 | 100 | `finished` |
| 102 | 2025-01-04 | C | A | 105 | 90 | `finished` |
| 103 | 2025-01-06 | A | D | 100 | 95 | `finished` |
| 104 | 2025-01-08 | A | B | null | null | `scheduled` |
| 105 | 2025-01-10 | C | A | 60 | 62 | `in_progress` |

### Resumo esperado do time A

| Métrica | Geral | Em casa | Fora |
|---|---|---|---|
| Jogos | 3 | 2 | 1 |
| Vitórias | 2 | 2 | 0 |
| Derrotas | 1 | 0 | 1 |
| Aproveitamento | 2/3 | 1 | 0 |
| Média de pontos a favor | 100 | 105 | 90 |
| Média de pontos contra | 100 | 97.5 | 105 |
| Saldo médio | 0 | 7.5 | -15 |

Cálculo geral dos pontos a favor:

`(110 + 90 + 100) / 3 = 100`

Cálculo geral dos pontos contra:

`(100 + 105 + 95) / 3 = 100`

As partidas 104 e 105 não entram nos cálculos.

### Últimos cinco jogos do time A

- Ordem esperada: 103, 102, 101.
- Jogos disponíveis: 3.
- As métricas coincidem com o resumo geral acima.

### Intervalo de 2025-01-04 a 2025-01-06

Os dois limites são inclusivos:

- Jogos considerados: 102 e 103.
- Jogos: 2.
- Vitórias: 1.
- Derrotas: 1.
- Aproveitamento: 0.5.
- Média de pontos a favor: 95.
- Média de pontos contra: 100.
- Saldo médio: -5.

### Intervalo de 2025-01-08 a 2025-01-10

Existem partidas nesse intervalo, mas nenhuma está encerrada:

- Jogos, vitórias e derrotas: 0.
- Aproveitamento e médias: `null`.