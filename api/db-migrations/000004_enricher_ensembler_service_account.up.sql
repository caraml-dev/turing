-- This update allows users to select the service account to be mounted in the
-- user-container in enricher and ensembler turing component
--
-- "service_account" column represents the secret name registered in the MLP project
-- that contains the service account JSON key value.

ALTER TABLE ensemblers
    ADD service_account text;

ALTER TABLE enrichers
    ADD service_account text;