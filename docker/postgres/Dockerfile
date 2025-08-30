FROM postgres:latest

# Create init scripts directory
RUN mkdir -p /docker-entrypoint-initdb.d

# Copy initialization scripts
# COPY init.sql /docker-entrypoint-initdb.d/

# Expose PostgreSQL port
EXPOSE 5432

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD pg_isready -U $POSTGRES_USER -d $POSTGRES_DB
