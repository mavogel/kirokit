# Warehouse Query

Guidance for writing and running queries against the analytics warehouse via `snowflake-mcp`.

1. Always run a `LIMIT`-bounded exploratory query before a full table scan.
2. Never run a `DELETE`, `DROP`, or `TRUNCATE` statement without explicit user confirmation.
3. Cite the source table and any filters applied when summarizing results.
