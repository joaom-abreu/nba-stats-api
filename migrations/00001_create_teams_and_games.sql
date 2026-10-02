-- +goose Up

CREATE TABLE teams (
                       id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                       source TEXT NOT NULL,
                       external_id TEXT NOT NULL,
                       name TEXT NOT NULL,
                       abbreviation TEXT NOT NULL,
                       conference TEXT NOT NULL,

                       CONSTRAINT teams_source_valid
                           CHECK (source IN ('demo', 'balldontlie')),

                       CONSTRAINT teams_external_id_present
                           CHECK (btrim(external_id) <> ''),

                       CONSTRAINT teams_name_present
                           CHECK (btrim(name) <> ''),

                       CONSTRAINT teams_abbreviation_present
                           CHECK (btrim(abbreviation) <> ''),

                       CONSTRAINT teams_conference_valid
                           CHECK (conference IN ('east', 'west')),

                       CONSTRAINT teams_external_identity_unique
                           UNIQUE (source, external_id),

                       CONSTRAINT teams_id_source_unique
                           UNIQUE (id, source)
);

CREATE TABLE games (
                       id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                       source TEXT NOT NULL,
                       external_id TEXT NOT NULL,
                       season INTEGER NOT NULL,
                       phase TEXT NOT NULL,
                       game_date DATE NOT NULL,
                       start_time TIMESTAMPTZ,
                       home_team_id BIGINT NOT NULL,
                       away_team_id BIGINT NOT NULL,
                       home_score INTEGER,
                       away_score INTEGER,
                       status TEXT NOT NULL,
                       collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                       CONSTRAINT games_source_valid
                           CHECK (source IN ('demo', 'balldontlie')),

                       CONSTRAINT games_external_id_present
                           CHECK (btrim(external_id) <> ''),

                       CONSTRAINT games_season_valid
                           CHECK (season > 0),

                       CONSTRAINT games_phase_valid
                           CHECK (phase IN ('regular_season', 'playoffs')),

                       CONSTRAINT games_status_valid
                           CHECK (
                               status IN (
                                          'scheduled',
                                          'in_progress',
                                          'finished',
                                          'postponed',
                                          'cancelled'
                                   )
                               ),

                       CONSTRAINT games_teams_different
                           CHECK (home_team_id <> away_team_id),

                       CONSTRAINT games_home_score_valid
                           CHECK (home_score >= 0),

                       CONSTRAINT games_away_score_valid
                           CHECK (away_score >= 0),

                       CONSTRAINT games_finished_result_valid
                           CHECK (
                               status <> 'finished'
                                   OR (
                                   home_score IS NOT NULL
                                       AND away_score IS NOT NULL
                                       AND home_score <> away_score
                                   )
                               ),

                       CONSTRAINT games_external_identity_unique
                           UNIQUE (source, external_id),

                       CONSTRAINT games_home_team_fk
                           FOREIGN KEY (home_team_id, source)
                               REFERENCES teams (id, source),

                       CONSTRAINT games_away_team_fk
                           FOREIGN KEY (away_team_id, source)
                               REFERENCES teams (id, source)
);

CREATE INDEX games_recorte_idx
    ON games (source, season, phase, game_date DESC, id DESC);

CREATE INDEX games_home_team_idx
    ON games (home_team_id);

CREATE INDEX games_away_team_idx
    ON games (away_team_id);

-- +goose Down

DROP TABLE games;
DROP TABLE teams;