SELECT 'teams' AS entity, COUNT(*) AS total
FROM teams
WHERE source = 'demo'

UNION ALL

SELECT 'games' AS entity, COUNT(*) AS total
FROM games
WHERE source = 'demo';

WITH team_games AS (
    SELECT
        CASE
            WHEN g.home_team_id = t.id THEN g.home_score
            ELSE g.away_score
            END AS points_for,

        CASE
            WHEN g.home_team_id = t.id THEN g.away_score
            ELSE g.home_score
            END AS points_against

    FROM games AS g
             JOIN teams AS t
                  ON t.id = g.home_team_id OR t.id = g.away_team_id

    WHERE t.source = 'demo'
      AND t.external_id = 'A'
      AND g.source = 'demo'
      AND g.season = 2024
      AND g.phase = 'regular_season'
      AND g.status = 'finished'
)
SELECT
    COUNT(*) AS games,
    COUNT(*) FILTER (WHERE points_for > points_against) AS wins,
    COUNT(*) FILTER (WHERE points_for < points_against) AS losses,
    ROUND(AVG(points_for), 2) AS avg_points_for,
    ROUND(AVG(points_against), 2) AS avg_points_against
FROM team_games;