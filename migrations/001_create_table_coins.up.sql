CREATE TABLE IF NOT EXISTS coins (
    id_coin serial,
	name varchar(256),
	price numeric(22,12),
	fixation_time bigint
);

CREATE INDEX idx_coin_timestamp ON coins (name, fixation_time);