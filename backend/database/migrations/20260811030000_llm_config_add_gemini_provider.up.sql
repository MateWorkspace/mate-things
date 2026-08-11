ALTER TABLE llm_config DROP CONSTRAINT chk_llm_config_provider;

ALTER TABLE llm_config
    ADD CONSTRAINT chk_llm_config_provider CHECK (provider IN ('CLAUDE', 'OPENAI', 'GEMINI'));
