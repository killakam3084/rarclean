#!/bin/sh
# Entrypoint wrapper: injects secrets via Infisical CLI then runs rarclean.
set -e

exec infisical run \
  --domain="${INFISICAL_DOMAIN:-https://infisical.iillmaticc.link}" \
  --projectId="${INFISICAL_PROJECT_ID}" \
  --env="${INFISICAL_ENV:-prod}" \
  -- ./rarclean "$@"
