#!/bin/sh
# Publishes this repository to its public GitHub mirror (ADR-0007).
#
# GitHub gets a fresh history: every run commits the public *tree* of the
# current commit on top of GitHub's main, with the same message. Bitbucket's
# own history is never pushed, because the commits before the public
# release contain unredacted Golden Files.
#
#   scripts/sync-github.sh               # on main: publish HEAD's public tree
#   scripts/sync-github.sh go/v0.2.0     # on a tag: tag the matching GitHub commit
#   scripts/sync-github.sh --export DIR  # write the public tree to DIR (for scanning)
#
# .publicignore lists the paths left out of the public tree, one per line
# (# starts a comment). A line starting with "!" puts a path back, so
# "docs/" followed by "!docs/api/" publishes docs/api/ only.
#
# Publishing needs push access to $GITHUB_REMOTE (CI uses a deploy key) and
# CYBERBIZ_REDACT_DENYLIST.
set -eu

remote=${GITHUB_REMOTE:-git@github.com:actinc/cyberbiz-sdk.git}

# public_tree prints the tree of HEAD as filtered by .publicignore.
public_tree() {
	[ -f .publicignore ] || { echo ".publicignore is missing; refusing to continue" >&2; exit 1; }
	index=$(mktemp)
	trap 'rm -f "$index"' EXIT
	GIT_INDEX_FILE=$index git read-tree HEAD
	grep -vE '^[[:space:]]*(#|$)' .publicignore | while read -r path; do
		case $path in
		!*) git ls-tree -r --full-tree HEAD -- "${path#!}" | GIT_INDEX_FILE=$index git update-index --index-info ;;
		*) GIT_INDEX_FILE=$index git rm -r -q --cached --ignore-unmatch -- "$path" ;;
		esac
	done
	GIT_INDEX_FILE=$index git write-tree
}
tree=$(public_tree)

if [ "${1:-}" = "--export" ]; then
	dir=${2:?usage: sync-github.sh --export DIR}
	mkdir -p "$dir"
	git archive "$tree" | tar -x -C "$dir"
	echo "exported the public tree $tree to $dir"
	exit 0
fi

# The mirror's main; empty on the very first run.
parent=
if git fetch --quiet "$remote" main 2>/dev/null; then
	parent=$(git rev-parse FETCH_HEAD)
fi

if [ $# -gt 0 ]; then
	tag=$1
	[ -n "$parent" ] || { echo "GitHub main is empty; sync main before tagging" >&2; exit 1; }
	target=$(git log --format='%H %T' "$parent" | awk -v t="$tree" '$2 == t { print $1; exit }')
	[ -n "$target" ] || { echo "no GitHub commit has the tree of $tag; sync main first" >&2; exit 1; }
	git push "$remote" "$target:refs/tags/$tag"
	echo "tagged $target as $tag"
	exit 0
fi

# Last line of defence: never publish a tree that names an integrator.
if [ -z "${CYBERBIZ_REDACT_DENYLIST:-}" ]; then
	echo "CYBERBIZ_REDACT_DENYLIST is not set; refusing to publish" >&2
	exit 1
fi
if git grep -qiE "$CYBERBIZ_REDACT_DENYLIST" "$tree" -- . ':!*.sum'; then
	git grep -ilE "$CYBERBIZ_REDACT_DENYLIST" "$tree" -- . ':!*.sum' >&2
	echo "files above contain denylisted names; refusing to publish" >&2
	exit 1
fi

if [ -n "$parent" ] && [ "$(git rev-parse "$parent^{tree}")" = "$tree" ]; then
	echo "GitHub main already has this tree"
	exit 0
fi

git config user.name "$(git log -1 --format=%an)"
git config user.email "$(git log -1 --format=%ae)"
if [ -n "$parent" ]; then
	commit=$(git log -1 --format=%B HEAD | git commit-tree "$tree" -p "$parent")
else
	commit=$(git log -1 --format=%B HEAD | git commit-tree "$tree")
fi
git push "$remote" "$commit:refs/heads/main"
echo "published $(git rev-parse --short HEAD) as $commit"
