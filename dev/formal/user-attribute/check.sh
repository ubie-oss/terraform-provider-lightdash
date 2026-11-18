#!/usr/bin/env bash
# Checks the TLA+ model of the lightdash_user_attribute update (UserAttributeUpdate.tla).
#
# 1. Conformance: Update still reads the attribute before it writes it back with the user
#    values that it read (exit 2 on drift).
# 2. TLC on sequential.cfg: the user value survives every apply when no UI edit runs during one.
#    It must pass. Exit 2 means the model's invariant no longer holds.
# 3. TLC on keep-users-removed.cfg: a regression that drops the user values must be caught.
#    It must fail. Exit 2 means the invariant cannot see a lost user value.
# 4. TLC on race.cfg: a UI edit between the read and the write is lost. This is the finding for
#    the current code. Exit 1 means the finding is present, and TLC's counterexample is printed.
#
# Exit codes: 0 = the model holds for the code, 1 = finding, 2 = drift or TLC failure.
# Needs Java 11 or later and rg. TLA2TOOLS=/path/to/tla2tools.jar skips the download.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "${HERE}/../../.." && pwd)"
TLC_VERSION="1.7.4"
TLC_SHA256="936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88"
SERVICE="${REPO}/internal/lightdash/services/user_attributes.go"
RESOURCE="${REPO}/internal/provider/resource_lightdash_user_attribute.go"

drift() {
	echo "DRIFT: $*" >&2
	exit 2
}

ensure_jar() {
	if [[ -n ${TLA2TOOLS:-} ]]; then
		JAR="${TLA2TOOLS}"
		return
	fi
	JAR="${XDG_CACHE_HOME:-${HOME}/.cache}/tla2tools-${TLC_VERSION}.jar"
	if [[ ! -f ${JAR} ]]; then
		mkdir -p "$(dirname "${JAR}")"
		curl -fsSL -o "${JAR}" "https://github.com/tlaplus/tlaplus/releases/download/v${TLC_VERSION}/tla2tools.jar"
	fi
	if ! echo "${TLC_SHA256}  ${JAR}" | shasum -a 256 -c - >/dev/null 2>&1; then
		rm -f "${JAR}"
		drift "tla2tools.jar checksum mismatch, removed the download"
	fi
}

# Prints a top-level func from its first line up to the closing brace at column 0.
# head must match the start of the func line exactly.
func_body() {
	awk -v head="$1" '
		index($0, head) == 1 { inside = 1 }
		inside { print }
		inside && /^}/ { exit }
	' "$2"
}

# Prints the line number of the first match of a fixed string within a captured body.
# Prints nothing and succeeds when there is no match, so the caller can report drift.
line_of() {
	{ rg -n -F -m 1 -- "$1" <<<"$2" || true; } | cut -d: -f1
}

check_conformance() {
	local update get_line keep_line put_line
	update="$(func_body 'func (s *UserAttributesService) Update(' "${SERVICE}")"
	[[ -n ${update} ]] || drift "UserAttributesService.Update not found in user_attributes.go"

	get_line="$(line_of 'GetByUUID(ctx, organizationUUID, attribute.UserAttributeUUID)' "${update}")"
	[[ -n ${get_line} ]] || drift "Update no longer reads the attribute with GetByUUID before it writes"

	keep_line="$(line_of 'keepUserValues(attribute, *current)' "${update}")"
	[[ -n ${keep_line} ]] || drift "Update no longer copies the users that it read into the write"

	put_line="$(line_of 'apiv1.UpdateUserAttributeV1(' "${update}")"
	[[ -n ${put_line} ]] || drift "Update no longer writes the attribute with apiv1.UpdateUserAttributeV1"

	if ((get_line >= keep_line || keep_line >= put_line)); then
		drift "Update no longer reads, copies, then writes the users in that order"
	fi

	local keep
	keep="$(func_body 'func keepUserValues(' "${SERVICE}")"
	rg -q -F 'desired.Users = current.Users' <<<"${keep}" ||
		drift "keepUserValues no longer copies the current users into the desired attribute"

	local resource_update
	resource_update="$(func_body 'func (r *userAttributeResource) Update(' "${RESOURCE}")"
	rg -q -F 'r.service.Update(ctx, desired.OrganizationUUID, desired)' <<<"${resource_update}" ||
		drift "the resource Update no longer calls the service Update"
}

# Runs TLC on one config. Prints the counterexample on a violation.
# Returns 0 when the invariant holds, 1 when TLC reports a violation, and exits 2 otherwise.
run_tlc() {
	local cfg="$1" meta out status=0
	meta="$(mktemp -d)"
	out="${meta}/tlc.out"
	java -XX:+UseParallelGC -cp "${JAR}" tlc2.TLC -workers 1 -metadir "${meta}" \
		-config "${HERE}/${cfg}" "${HERE}/UserAttributeUpdate.tla" >"${out}" 2>&1 || status=$?
	if grep -q "No error has been found" "${out}"; then
		echo "PASS ${cfg}"
		return 0
	fi
	if grep -q "is violated" "${out}"; then
		grep -E 'is violated|^State [0-9]+|^/\\|^Error: The behavior' "${out}" | head -40
		echo "VIOLATED ${cfg} (TLC output: ${out})"
		return 1
	fi
	cat "${out}" >&2
	drift "TLC did not finish for ${cfg} (exit ${status})"
}

ensure_jar
check_conformance
echo "conformance: Update reads, copies the users, then writes"

status=0
run_tlc sequential.cfg || status=$?
if [[ ${status} -ne 0 ]]; then
	drift "sequential.cfg fails: a user value is lost without any UI edit during an apply"
fi

status=0
run_tlc keep-users-removed.cfg || status=$?
if [[ ${status} -ne 1 ]]; then
	drift "keep-users-removed.cfg passes: the invariant cannot see a regression that drops the users"
fi
echo "self-test: the invariant catches a regression that drops the users"

status=0
run_tlc race.cfg || status=$?
if [[ ${status} -eq 1 ]]; then
	echo "FINDING: a UI edit between Get and Put is lost (race.cfg counterexample above)"
	exit 1
fi
drift "race.cfg passes: the model no longer finds the read-then-write window"
