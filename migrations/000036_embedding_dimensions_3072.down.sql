DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM memory_chunks WHERE embedding IS NOT NULL LIMIT 1)
        OR EXISTS (SELECT 1 FROM embedding_cache WHERE embedding IS NOT NULL LIMIT 1)
        OR EXISTS (SELECT 1 FROM skills WHERE embedding IS NOT NULL LIMIT 1)
        OR EXISTS (SELECT 1 FROM agents WHERE embedding IS NOT NULL LIMIT 1)
        OR EXISTS (SELECT 1 FROM team_tasks WHERE embedding IS NOT NULL LIMIT 1)
        OR EXISTS (SELECT 1 FROM kg_entities WHERE embedding IS NOT NULL LIMIT 1) THEN
        RAISE EXCEPTION 'migration 000036 rollback requires all embedding columns to be empty before changing vector dimensions';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_mem_vec;
DROP INDEX IF EXISTS idx_skills_embedding;
DROP INDEX IF EXISTS idx_agents_embedding;
DROP INDEX IF EXISTS idx_tt_embedding;
DROP INDEX IF EXISTS idx_kg_entity_vec;

ALTER TABLE memory_chunks ALTER COLUMN embedding TYPE vector(1536);
ALTER TABLE embedding_cache ALTER COLUMN embedding TYPE vector(1536);
ALTER TABLE skills ALTER COLUMN embedding TYPE vector(1536);
ALTER TABLE agents ALTER COLUMN embedding TYPE vector(1536);
ALTER TABLE team_tasks ALTER COLUMN embedding TYPE vector(1536);
ALTER TABLE kg_entities ALTER COLUMN embedding TYPE vector(1536);

CREATE INDEX idx_mem_vec ON memory_chunks USING hnsw(embedding vector_cosine_ops);
CREATE INDEX idx_skills_embedding ON skills USING hnsw(embedding vector_cosine_ops);
CREATE INDEX idx_agents_embedding ON agents USING hnsw(embedding vector_cosine_ops);
CREATE INDEX idx_tt_embedding ON team_tasks USING hnsw (embedding vector_cosine_ops);
CREATE INDEX idx_kg_entity_vec ON kg_entities USING hnsw(embedding vector_cosine_ops);
