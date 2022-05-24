-- Introduce Python version column, with default as version 3.7 which is the
-- only supported major version as of the introduction of the column.
ALTER TABLE pyfunc_ensemblers ADD python_version varchar(16) DEFAULT '3.7.*';
