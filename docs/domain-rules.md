# Regras de domínio e métricas

Este documento define o comportamento esperado da NBA Stats API.

## 1. Temporada, fase e origem

- A temporada é identificada pelo ano em que começa.
- A temporada `2024` pode conter partidas realizadas em 2025.
- A fase faz parte do recorte e não é deduzida pela data da partida.
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