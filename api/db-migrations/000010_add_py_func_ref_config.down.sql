-- Removes py_func_ref_config from table.
-- Hence, this migration involves data loss for ensemblers with type that is "pyfunc"
ALTER TABLE ensembler_configs DROP COLUMN py_func_ref_config;

