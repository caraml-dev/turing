-- Removes pyfunc_config from table.
-- Hence, this migration involves data loss for ensemblers with type that is "pyfunc"
ALTER TABLE ensembler_configs DROP COLUMN pyfunc_config;
