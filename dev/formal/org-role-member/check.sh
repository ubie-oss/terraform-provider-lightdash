#!/usr/bin/env bash
# Checks the TLA+ model of the lightdash_organization_role_member lifecycle.
#
# 1. Conformance: the code still matches the model's assumptions (exit 2 on drift).
# 2. TLC on fixture.cfg: the model holds under the catalog that the tests assume.
# 3. TLC on the config that matches the code: live.cfg when role names resolve through
#    the catalog, slug.cfg otherwise. Exit 1 means the code breaks NoRoleNotFound.
#
# Exit codes: 0 = the model holds for the code, 1 = finding, 2 = drift or TLC failure.
# Needs Java 11 or later and rg. TLA2TOOLS=/path/to/tla2tools.jar skips the download.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$HERE/../../.." && pwd)"
TLC_VERSION="1.7.4"
TLC_SHA256="936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88"
RESOURCE="$REPO/internal/provider/resource_lightdash_organization_role_member.go"
SERVICE="$REPO/internal/lightdash/services/roles.go"

drift() {
	echo "DRIFT: $*" >&2
	exit 2
}

ensure_jar() {
	if [ -n "${TLA2TOOLS:-}" ]; then
		JAR="$TLA2TOOLS"
		return
	fi
	JAR="${XDG_CACHE_HOME:-$HOME/.cache}/tla2tools-$TLC_VERSION.jar"
	if [ ! -f "$JAR" ]; then
		mkdir -p "$(dirname "$JAR")"
		curl -fsSL -o "$JAR" "https://github.com/tlaplus/tlaplus/releases/download/v$TLC_VERSION/tla2tools.jar"
	fi
	if ! echo "$TLC_SHA256  $JAR" | shasum -a 256 -c - >/dev/null 2>&1; then
		rm -f "$JAR"
		drift "tla2tools.jar checksum mismatch, removed the download"
	fi
}

check_conformance() {
	rg -U -q 'func \(r \*organizationRoleMemberResource\) Create\([\s\S]*?AssignOrgUserRole\(ctx, orgUUID, userUUID, role\.String\(\)\)' "$RESOURCE" ||
		drift "Create no longer assigns the planned role through AssignOrgUserRole"
	rg -U -q 'func \(r \*organizationRoleMemberResource\) Update\([\s\S]*?AssignOrgUserRole\(ctx, orgUUID, userUUID, role\.String\(\)\)' "$RESOURCE" ||
		drift "Update no longer assigns the planned role through AssignOrgUserRole"
	rg -U -q 'func \(r \*organizationRoleMemberResource\) Delete\([\s\S]*?AssignOrgUserRole\(ctx, orgUUID, userUUID, models\.ORGANIZATION_MEMBER_ROLE\.String\(\)\)' "$RESOURCE" ||
		drift "Delete no longer resets the user to the default role member"
}

code_mode() {
	if rg -U -q 'func \(s \*RoleService\) AssignOrgUserRole\([^)]*\) \(\*models\.RoleAssignment, error\) \{\s*roleID, err := s\.ResolveRoleID' "$SERVICE"; then
		echo catalog
	else
		echo slug
	fi
}

run_tlc() {
	local cfg="$1" meta out status
	meta="$(mktemp -d)"
	out="$meta/tlc.out"
	set +e
	java -XX:+UseParallelGC -cp "$JAR" tlc2.TLC -workers 1 -metadir "$meta" \
		-config "$HERE/$cfg" "$HERE/OrgRoleMember.tla" >"$out" 2>&1
	status=$?
	set -e
	if grep -q "No error has been found" "$out"; then
		echo "PASS $cfg"
		return 0
	fi
	if grep -q "is violated" "$out"; then
		grep -E 'is violated|^State [0-9]+|^/\\|^Error: The behavior' "$out" | head -40
		echo "FINDING $cfg (TLC output: $out)"
		return 1
	fi
	cat "$out" >&2
	drift "TLC did not finish for $cfg (exit $status)"
}

ensure_jar
check_conformance
MODE="$(code_mode)"
echo "code resolves org roles: $MODE"

run_tlc fixture.cfg || drift "fixture.cfg fails: the invariants no longer hold under the test-fixture catalog"

if [ "$MODE" = catalog ]; then
	set +e
	run_tlc live.cfg
	status=$?
	set -e
	exit "$status"
fi
run_tlc slug.cfg
