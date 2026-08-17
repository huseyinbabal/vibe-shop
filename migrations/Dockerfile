# Migration runner image: SQL files baked in, applied once each in filename
# order (tracked in schema_migrations) so re-deploys stay idempotent.
FROM postgres:16-alpine
COPY *.sql /migrations/
COPY apply.sh /apply.sh
ENTRYPOINT ["/bin/sh", "/apply.sh"]
