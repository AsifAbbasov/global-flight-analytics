#!/usr/bin/env bash
set -euo pipefail

container_id="${1:-}"
database_user="${2:-}"
database_name="${3:-}"
maximum_attempts="${4:-60}"
retry_delay_seconds="${5:-1}"

if [ -z "$container_id" ] ||
   [ -z "$database_user" ] ||
   [ -z "$database_name" ]; then
  printf '%s\n' \
    "usage: wait-for-postgres-container.sh <container> <user> <database> [attempts] [delay-seconds]" \
    >&2
  exit 2
fi

case "$maximum_attempts" in
  ''|*[!0-9]*)
    printf '%s\n' "attempts must be a positive integer" >&2
    exit 2
    ;;
esac

case "$retry_delay_seconds" in
  ''|*[!0-9]*)
    printf '%s\n' "delay-seconds must be a non-negative integer" >&2
    exit 2
    ;;
esac

if [ "$maximum_attempts" -le 0 ]; then
  printf '%s\n' "attempts must be greater than zero" >&2
  exit 2
fi

for attempt in $(seq 1 "$maximum_attempts"); do
  main_process="$(
    docker exec \
      "$container_id" \
      cat /proc/1/comm \
      2>/dev/null ||
      true
  )"

  if [ "$main_process" = "postgres" ]; then
    query_result="$(
      docker exec \
        "$container_id" \
        psql \
        --username "$database_user" \
        --dbname "$database_name" \
        --no-align \
        --tuples-only \
        --command 'SELECT 1' \
        2>/dev/null ||
        true
    )"

    normalized_result="$(
      printf '%s' "$query_result" |
        tr -d '[:space:]'
    )"

    if [ "$normalized_result" = "1" ]; then
      printf '%s\n' \
        "PostgreSQL container is ready: container=$container_id database=$database_name attempt=$attempt"
      exit 0
    fi
  fi

  sleep "$retry_delay_seconds"
done

printf '%s\n' \
  "PostgreSQL container did not reach final database readiness: container=$container_id database=$database_name attempts=$maximum_attempts" \
  >&2

docker logs "$container_id" >&2 || true
exit 1
