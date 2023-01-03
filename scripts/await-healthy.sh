#!/usr/bin/env sh

set -e

printf 'Waiting for GitLab container to become healthy'

until test -n "$(docker ps --quiet --filter label=terraform-provider-gitlab/owned --filter health=healthy)"; do
  printf '.'
  sleep 5
done

echo
echo "GitLab is healthy at $GITLAB_BASE_URL"

# Print the version, since it is useful debugging information.
curl --silent --show-error --header "Authorization: Bearer $GITLAB_TOKEN" "$GITLAB_BASE_URL/version"
echo
