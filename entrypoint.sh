#!/bin/sh
set -u -e

NOW=$(date "+%Y-%m-%d %T")
PAGES="$1"
OUTPUT="/app/repo/$2"

if [ $# -lt 2 ]; then
  echo 1>&2 "$0: not enough arguments, provide <pages> <output-dir>"
  exit 2
fi

# Use this user email https://github.com/actions
git config --global user.email 41898282+github-actions[bot]@users.noreply.github.com
git config --global user.name "Notion Exporter"

echo ":: Cloning github.com/${GITHUB_REPOSITORY} into /app/repo"
git clone "https://${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}" /app/repo

echo ":: Exporting ${PAGES} into /app/repo"
/app/notion-exporter -pages "${PAGES}" -token "${NOTION_TOKEN}" -output "$OUTPUT"

find "${OUTPUT}"

set +e
cd /app/repo && git add "$2" && git commit -am "Pulled content from notion at ${NOW}"
set -e

if [ "$?" -ne "0" ]; then
    echo "nothing to commit"
    exit 0
else
  git push
fi


