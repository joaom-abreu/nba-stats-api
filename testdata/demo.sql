BEGIN;

INSERT INTO teams (
    source, external_id, name, abbreviation, conference
)
VALUES
    ('demo', 'A', 'Demo Team A', 'A', 'east'),
    ('demo', 'B', 'Demo Team B', 'B', 'east'),
    ('demo', 'C', 'Demo Team C', 'C', 'west'),
    ('demo', 'D', 'Demo Team D', 'D', 'west')
    ON CONFLICT (source, external_id)
DO UPDATE SET
    name = EXCLUDED.name,
           abbreviation = EXCLUDED.abbreviation,
           conference = EXCLUDED.conference;

WITH demo_games (
                 external_id,
                 game_date,
                 home_external_id,
                 away_external_id,
                 home_score,
                 away_score,
                 status
    ) AS (
    VALUES
        ('101', DATE '2025-01-02', 'A', 'B', 110, 100, 'finished'),
        ('102', DATE '2025-01-04', 'C', 'A', 105, 90, 'finished'),
        ('103', DATE '2025-01-06', 'A', 'D', 100, 95, 'finished'),
        ('104', DATE '2025-01-08', 'A', 'B', NULL, NULL, 'scheduled'),
        ('105', DATE '2025-01-10', 'C', 'A', 60, 62, 'in_progress')
)
INSERT INTO games (
    source,
    external_id,
    season,
    phase,
    game_date,
    start_time,
    home_team_id,
    away_team_id,
    home_score,
    away_score,
    status,
    collected_at
)
SELECT
    'demo',
    g.external_id,
    2024,
    'regular_season',
    g.game_date,
    NULL,
    (
        SELECT id
        FROM teams
        WHERE source = 'demo'
          AND external_id = g.home_external_id
    ),
    (
        SELECT id
        FROM teams
        WHERE source = 'demo'
          AND external_id = g.away_external_id
    ),
    g.home_score,
    g.away_score,
    g.status,
    now()
FROM demo_games AS g
WHERE true
    ON CONFLICT (source, external_id)
DO UPDATE SET
    season = EXCLUDED.season,
           phase = EXCLUDED.phase,
           game_date = EXCLUDED.game_date,
           start_time = EXCLUDED.start_time,
           home_team_id = EXCLUDED.home_team_id,
           away_team_id = EXCLUDED.away_team_id,
           home_score = EXCLUDED.home_score,
           away_score = EXCLUDED.away_score,
           status = EXCLUDED.status,
           collected_at = EXCLUDED.collected_at;

COMMIT;