import json
import logging
import asyncpg
from typing import List, Dict, Any, Optional
from app.config import settings

logger = logging.getLogger("chatbot.database")

class Database:
    pool: Optional[asyncpg.Pool] = None

db = Database()


async def init_db() -> None:
    """Initialize asyncpg connection pool."""
    logger.info(f"Connecting to PostgreSQL at {settings.database_url}")
    try:
        db.pool = await asyncpg.create_pool(
            dsn=settings.database_url,
            min_size=settings.db_min_size,
            max_size=settings.db_max_size,
            command_timeout=60,
        )
        logger.info("PostgreSQL database pool connected successfully!")
    except Exception as e:
        logger.error(f"Failed to create PostgreSQL database pool: {e}")
        raise e


async def close_db() -> None:
    """Close asyncpg connection pool."""
    if db.pool:
        logger.info("Closing PostgreSQL connection pool...")
        await db.pool.close()
        logger.info("PostgreSQL connection pool closed.")


async def search_similar_chunks(
    embedding: List[float], 
    top_k: int = 5, 
    collection_name: str = "aria_knowledge"
) -> List[Dict[str, Any]]:
    """
    Perform cosine similarity vector search using pgvector (<=> operator).
    Searches langchain_pg_embedding table associated with the target collection.
    """
    if not db.pool:
        raise RuntimeError("Database connection pool is not initialized.")

    vector_str = f"[{','.join(map(str, embedding))}]"

    # Query connecting langchain_pg_embedding and langchain_pg_collection
    query_with_collection = """
        SELECT 
            e.document AS content,
            e.cmetadata AS metadata,
            1 - (e.embedding <=> $1::vector) AS similarity
        FROM langchain_pg_embedding e
        JOIN langchain_pg_collection c ON c.uuid = e.collection_id
        WHERE c.name = $2
        ORDER BY e.embedding <=> $1::vector
        LIMIT $3;
    """

    # Fallback query if collection table join fails or isn't used
    fallback_query = """
        SELECT 
            document AS content,
            cmetadata AS metadata,
            1 - (embedding <=> $1::vector) AS similarity
        FROM langchain_pg_embedding
        ORDER BY embedding <=> $1::vector
        LIMIT $2;
    """

    async with db.pool.acquire() as conn:
        try:
            rows = await conn.fetch(query_with_collection, vector_str, collection_name, top_k)
        except Exception as err:
            logger.warning(f"Collection join query failed ({err}), trying direct embedding table query...")
            rows = await conn.fetch(fallback_query, vector_str, top_k)

        results = []
        for row in rows:
            metadata_raw = row["metadata"]
            if isinstance(metadata_raw, str):
                try:
                    metadata_dict = json.loads(metadata_raw)
                except json.JSONDecodeError:
                    metadata_dict = {"raw": metadata_raw}
            elif isinstance(metadata_raw, dict):
                metadata_dict = metadata_raw
            else:
                metadata_dict = {}

            results.append({
                "content": row["content"] or "",
                "metadata": metadata_dict,
                "similarity": float(row["similarity"]) if row["similarity"] is not None else 0.0,
            })

        return results
