-- Executado automaticamente na primeira subida do volume do TimescaleDB.
CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS sensor_readings (
    time      TIMESTAMPTZ      NOT NULL,
    sensor_id TEXT             NOT NULL,
    value     DOUBLE PRECISION NOT NULL
);

-- Converte a tabela em hypertable particionada pelo tempo.
SELECT create_hypertable('sensor_readings', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_sensor_readings_sensor_time
    ON sensor_readings (sensor_id, time DESC);
