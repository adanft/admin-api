#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_URL="${API_URL:-http://localhost:9090}"
AUTO_START="${AUTO_START:-1}"
RESET_DB="${RESET_DB:-0}"
WAIT_SECONDS="${WAIT_SECONDS:-90}"

log() {
  printf '[e2e] %s\n' "$*"
}

fail() {
  printf '[e2e][FAIL] %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "Required command not found: $1"
}

api_ready() {
  local code
  code="$(curl -s -o /dev/null -w '%{http_code}' "${API_URL}/users" || true)"
  [[ "$code" == "200" ]]
}

start_stack_if_needed() {
  if [[ "$AUTO_START" != "1" ]]; then
    return
  fi

  if api_ready; then
    log "API already reachable at ${API_URL}"
    return
  fi

  require_cmd docker
  if [[ ! -f "${ROOT_DIR}/docker-compose.yml" ]]; then
    fail "docker-compose.yml not found in ${ROOT_DIR}"
  fi

  if [[ "$RESET_DB" == "1" ]]; then
    log "Resetting stack (docker compose down -v)"
    (cd "$ROOT_DIR" && docker compose down -v)
  fi

  log "Starting stack with docker compose"
  (cd "$ROOT_DIR" && docker compose up -d)
}

wait_api() {
  local i
  for ((i = 1; i <= WAIT_SECONDS; i++)); do
    if api_ready; then
      log "API ready after ${i}s"
      return
    fi
    sleep 1
  done

  fail "API not reachable at ${API_URL} after ${WAIT_SECONDS}s"
}

request() {
  local method="$1"
  local path="$2"
  local payload="${3:-}"
  local content_type="${4:-application/json}"

  local tmp status
  tmp="$(mktemp)"

  if [[ -n "$payload" ]]; then
    status="$(curl -sS -o "$tmp" -w '%{http_code}' -X "$method" \
      -H "Content-Type: ${content_type}" \
      --data "$payload" \
      "${API_URL}${path}" || true)"
  else
    status="$(curl -sS -o "$tmp" -w '%{http_code}' -X "$method" \
      "${API_URL}${path}" || true)"
  fi

  RESPONSE_STATUS="$status"
  RESPONSE_BODY="$(cat "$tmp")"
  rm -f "$tmp"
}

request_file() {
  local method="$1"
  local path="$2"
  local file="$3"
  local content_type="${4:-application/json}"

  local tmp status
  tmp="$(mktemp)"

  status="$(curl -sS -o "$tmp" -w '%{http_code}' -X "$method" \
    -H "Content-Type: ${content_type}" \
    --data-binary "@${file}" \
    "${API_URL}${path}" || true)"

  RESPONSE_STATUS="$status"
  RESPONSE_BODY="$(cat "$tmp")"
  rm -f "$tmp"
}

random_uuid() {
  if [[ -r /proc/sys/kernel/random/uuid ]]; then
    cat /proc/sys/kernel/random/uuid
    return
  fi
  if command -v uuidgen >/dev/null 2>&1; then
    uuidgen | tr '[:upper:]' '[:lower:]'
    return
  fi

  # Stable fallback in case UUID generators are unavailable.
  echo "00000000-0000-0000-0000-000000000123"
}

delete_user_if_exists() {
  local id="$1"
  request "DELETE" "/users/${id}"
  if [[ "$RESPONSE_STATUS" != "204" && "$RESPONSE_STATUS" != "404" ]]; then
    fail "cleanup failed for user ${id}: http=${RESPONSE_STATUS} body=${RESPONSE_BODY}"
  fi
}

assert_status() {
  local expected="$1"
  local label="$2"
  if [[ "$RESPONSE_STATUS" != "$expected" ]]; then
    fail "${label}: expected HTTP ${expected}, got ${RESPONSE_STATUS}. body=${RESPONSE_BODY}"
  fi
}

assert_jq() {
  local expr="$1"
  local label="$2"
  if ! jq -e "$expr" >/dev/null 2>&1 <<<"$RESPONSE_BODY"; then
    fail "${label}: jq assertion failed (${expr}). body=${RESPONSE_BODY}"
  fi
}

run_tests() {
  local suffix username username_upd email updated_email upper_email user_id
  local missing_user_id large_payload_file

  suffix="$(date +%s)$RANDOM"
  username="u_${suffix}"
  username_upd="${username}u"
  email="user_${suffix}@example.com"
  updated_email="upd_${email}"
  upper_email="$(tr '[:lower:]' '[:upper:]' <<<"$updated_email")"

  log "T1: GET /users"
  request "GET" "/users"
  assert_status "200" "T1"
  assert_jq '.success == true and (.data | type == "array")' "T1"

  log "T2: POST /users with invalid content-type"
  request "POST" "/users" '{"name":"A"}' "text/plain"
  assert_status "400" "T2"
  assert_jq '.success == false and .code == "INVALID_CONTENT_TYPE"' "T2"

  log "T3: POST /users valid"
  request "POST" "/users" "{\"name\":\"Admin\",\"lastName\":\"User\",\"username\":\"${username}\",\"email\":\"${email}\",\"avatar\":\"https://example.com/a.png\"}"
  assert_status "201" "T3"
  assert_jq '.success == true and (.data.id | length > 0)' "T3"
  assert_jq '.data.username == "'"${username}"'" and .data.email == "'"${email}"'"' "T3"
  user_id="$(jq -r '.data.id' <<<"$RESPONSE_BODY")"
  if [[ -z "$user_id" || "$user_id" == "null" ]]; then
    fail "T3: invalid user id extracted"
  fi

  log "T4: GET /users/{id}"
  request "GET" "/users/${user_id}"
  assert_status "200" "T4"
  assert_jq '.success == true and .data.id == "'"${user_id}"'"' "T4"

  log "T5: PUT /users/{id}"
  request "PUT" "/users/${user_id}" "{\"name\":\"Admin2\",\"lastName\":\"User2\",\"username\":\"${username_upd}\",\"email\":\"${updated_email}\",\"avatar\":\"https://example.com/b.png\"}"
  assert_status "200" "T5"
  assert_jq '.success == true and .data.username == "'"${username_upd}"'"' "T5"

  log "T6: GET /users list contains user"
  request "GET" "/users"
  assert_status "200" "T6"
  assert_jq '.success == true and ([.data[].id] | index("'"${user_id}"'") != null)' "T6"

  log "T7: POST duplicate email (case-insensitive)"
  request "POST" "/users" "{\"name\":\"Dup\",\"lastName\":\"Mail\",\"username\":\"${username}_dup\",\"email\":\"${upper_email}\",\"avatar\":\"\"}"
  assert_status "409" "T7"
  assert_jq '.success == false and .code == "EMAIL_EXISTS"' "T7"

  log "T8: PUT invalid id"
  request "PUT" "/users/not-a-uuid" '{"name":"x","lastName":"y","username":"z","email":"z@example.com","avatar":""}'
  assert_status "400" "T8"
  assert_jq '.success == false and .code == "INVALID_ID"' "T8"

  log "T9: POST /users required fields empty"
  request "POST" "/users" '{"name":"","lastName":"","username":"","email":"","avatar":""}'
  assert_status "400" "T9"
  assert_jq '.success == false and .code == "INVALID_PAYLOAD"' "T9"

  log "T10: POST /users invalid email"
  request "POST" "/users" "{\"name\":\"Bad\",\"lastName\":\"Email\",\"username\":\"bad_${suffix}\",\"email\":\"invalid-email\",\"avatar\":\"\"}"
  assert_status "400" "T10"
  assert_jq '.success == false and .code == "INVALID_EMAIL"' "T10"

  log "T11: POST /users duplicate username"
  request "POST" "/users" "{\"name\":\"Dup\",\"lastName\":\"User\",\"username\":\"${username_upd}\",\"email\":\"dupu_${email}\",\"avatar\":\"\"}"
  assert_status "409" "T11"
  assert_jq '.success == false and .code == "USERNAME_EXISTS"' "T11"

  log "T12: POST /users malformed JSON"
  request "POST" "/users" '{"name":"broken",'
  assert_status "400" "T12"
  assert_jq '.success == false and .code == "MALFORMED_JSON"' "T12"

  log "T13: POST /users invalid field type"
  request "POST" "/users" "{\"name\":\"Type\",\"lastName\":\"Error\",\"username\":\"type_${suffix}\",\"email\":\"type_${email}\",\"avatar\":123}"
  assert_status "400" "T13"
  assert_jq '.success == false and .code == "INVALID_FIELD_TYPE"' "T13"

  log "T14: POST /users multiple JSON objects"
  request "POST" "/users" '{"name":"A","lastName":"B","username":"multi1","email":"multi1@example.com","avatar":""}{"name":"C"}'
  assert_status "400" "T14"
  assert_jq '.success == false and .code == "MULTIPLE_JSON_OBJECTS"' "T14"

  log "T15: POST /users unknown field"
  request "POST" "/users" "{\"name\":\"Unknown\",\"lastName\":\"Field\",\"username\":\"unk_${suffix}\",\"email\":\"unk_${email}\",\"avatar\":\"\",\"unexpected\":true}"
  assert_status "400" "T15"
  assert_jq '.success == false and .code == "INVALID_PAYLOAD"' "T15"

  log "T16: POST /users body too large"
  large_payload_file="$(mktemp)"
  {
    printf '{"name":"Big","lastName":"Payload","username":"u_big_%s","email":"big_%s@example.com","avatar":"' "$suffix" "$suffix"
    head -c 1100000 </dev/zero | tr '\0' 'a'
    printf '"}'
  } >"${large_payload_file}"
  request_file "POST" "/users" "${large_payload_file}"
  rm -f "${large_payload_file}"
  assert_status "400" "T16"
  assert_jq '.success == false and .code == "BODY_TOO_LARGE"' "T16"

  log "T17: GET /users invalid id"
  request "GET" "/users/not-a-uuid"
  assert_status "400" "T17"
  assert_jq '.success == false and .code == "INVALID_ID"' "T17"

  log "T18: DELETE /users invalid id"
  request "DELETE" "/users/not-a-uuid"
  assert_status "400" "T18"
  assert_jq '.success == false and .code == "INVALID_ID"' "T18"

  log "T19: PUT /users/{id} not found"
  missing_user_id="$(random_uuid)"
  if [[ "$missing_user_id" == "$user_id" ]]; then
    missing_user_id="00000000-0000-0000-0000-000000000124"
  fi
  request "PUT" "/users/${missing_user_id}" '{"name":"Missing","lastName":"User","username":"missing_user","email":"missing@example.com","avatar":""}'
  assert_status "404" "T19"
  assert_jq '.success == false and .code == "NOT_FOUND"' "T19"

  log "T22: DELETE /users/{id}"
  request "DELETE" "/users/${user_id}"
  assert_status "204" "T22"

  log "T23: GET deleted user"
  request "GET" "/users/${user_id}"
  assert_status "404" "T23"
  assert_jq '.success == false and .code == "NOT_FOUND"' "T23"

  log "All endpoint tests passed."
}

main() {
  require_cmd curl
  require_cmd jq

  start_stack_if_needed
  wait_api
  run_tests
}

main "$@"
