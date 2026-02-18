#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_URL="${API_URL:-http://localhost:9090}"
AUTO_START="${AUTO_START:-1}"
RESET_DB="${RESET_DB:-0}"
WAIT_SECONDS="${WAIT_SECONDS:-90}"
REFRESH_COOKIE_NAME="${AUTH_REFRESH_COOKIE_NAME:-refresh_token}"
RUN_EXPIRY_TESTS="${RUN_EXPIRY_TESTS:-1}"
MAX_WAIT_ACCESS_EXP_SECONDS="${MAX_WAIT_ACCESS_EXP_SECONDS:-180}"
MAX_WAIT_REFRESH_EXP_SECONDS="${MAX_WAIT_REFRESH_EXP_SECONDS:-180}"
FORCE_EXPIRY_WAIT="${FORCE_EXPIRY_WAIT:-0}"

COOKIE_JAR=""
RESPONSE_STATUS=""
RESPONSE_BODY=""
RESPONSE_HEADERS=""

log() {
  printf '[auth-e2e] %s\n' "$*"
}

fail() {
  printf '[auth-e2e][FAIL] %s\n' "$*" >&2
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
  [[ -f "${ROOT_DIR}/docker-compose.yml" ]] || fail "docker-compose.yml not found in ${ROOT_DIR}"

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
  local auth_header="${5:-}"
  local with_cookie="${6:-0}"

  local body_tmp header_tmp status
  body_tmp="$(mktemp)"
  header_tmp="$(mktemp)"

  local curl_args
  curl_args=(-sS -D "$header_tmp" -o "$body_tmp" -w '%{http_code}' -X "$method")

  if [[ -n "$content_type" ]]; then
    curl_args+=(-H "Content-Type: ${content_type}")
  fi
  if [[ -n "$auth_header" ]]; then
    curl_args+=(-H "Authorization: ${auth_header}")
  fi
  if [[ "$with_cookie" == "1" ]]; then
    curl_args+=(-b "$COOKIE_JAR" -c "$COOKIE_JAR")
  fi
  if [[ -n "$payload" ]]; then
    curl_args+=(--data "$payload")
  fi
  curl_args+=("${API_URL}${path}")

  status="$(curl "${curl_args[@]}" || true)"

  RESPONSE_STATUS="$status"
  RESPONSE_BODY="$(<"$body_tmp")"
  RESPONSE_HEADERS="$(<"$header_tmp")"
  rm -f "$body_tmp" "$header_tmp"
}

request_with_cookie_value() {
  local method="$1"
  local path="$2"
  local cookie_name="$3"
  local cookie_value="$4"

  local body_tmp header_tmp status
  body_tmp="$(mktemp)"
  header_tmp="$(mktemp)"

  status="$(curl -sS -D "$header_tmp" -o "$body_tmp" -w '%{http_code}' -X "$method" -H "Cookie: ${cookie_name}=${cookie_value}" "${API_URL}${path}" || true)"

  RESPONSE_STATUS="$status"
  RESPONSE_BODY="$(<"$body_tmp")"
  RESPONSE_HEADERS="$(<"$header_tmp")"
  rm -f "$body_tmp" "$header_tmp"
}

assert_status() {
  local expected="$1"
  local label="$2"
  [[ "$RESPONSE_STATUS" == "$expected" ]] || fail "${label}: expected HTTP ${expected}, got ${RESPONSE_STATUS}. body=${RESPONSE_BODY}"
}

assert_jq() {
  local expr="$1"
  local label="$2"
  jq -e "$expr" >/dev/null 2>&1 <<<"$RESPONSE_BODY" || fail "${label}: jq assertion failed (${expr}). body=${RESPONSE_BODY}"
}

assert_cookie_header() {
  local cookie_name="$1"
  local label="$2"
  grep -q "$cookie_name" <<<"$RESPONSE_HEADERS" || fail "${label}: expected Set-Cookie for ${cookie_name}. headers=${RESPONSE_HEADERS}"
}

cookie_token_from_headers() {
  local headers="$1"
  local cookie_name="$2"
  sed -n "s/^Set-Cookie: ${cookie_name}=\([^;]*\).*/\1/pI" <<<"$headers" | head -n 1
}

base64url_decode() {
  local input="$1"
  local normalized padded

  normalized="$(tr '_-' '/+' <<<"$input")"
  case $((${#normalized} % 4)) in
    2) padded="${normalized}==" ;;
    3) padded="${normalized}=" ;;
    0) padded="$normalized" ;;
    *) padded="$normalized" ;;
  esac

  base64 -d <<<"$padded" 2>/dev/null || return 1
}

token_exp_epoch() {
  local token="$1"
  local payload_part payload_json

  payload_part="$(cut -d'.' -f2 <<<"$token")"
  payload_json="$(base64url_decode "$payload_part")" || return 1
  jq -r '.exp // 0' <<<"$payload_json"
}

cookie_exp_epoch() {
  local headers="$1"
  local cookie_name="$2"
  local line expires

  line="$(grep -i "^Set-Cookie: ${cookie_name}=" <<<"$headers" | head -n 1 || true)"
  [[ -n "$line" ]] || return 1

  expires="$(sed -n 's/.*[eE]xpires=\([^;]*\).*/\1/p' <<<"$line")"
  [[ -n "$expires" ]] || return 1

  date -d "$expires" +%s 2>/dev/null
}

wait_until_epoch_or_skip() {
  local label="$1"
  local target_epoch="$2"
  local max_wait="$3"
  local now wait_seconds

  now="$(date +%s)"
  wait_seconds=$((target_epoch - now + 1))

  if ((wait_seconds <= 0)); then
    return 0
  fi

  if ((wait_seconds > max_wait)) && [[ "$FORCE_EXPIRY_WAIT" != "1" ]]; then
    log "${label}: skipped (wait=${wait_seconds}s > max=${max_wait}s). Usa FORCE_EXPIRY_WAIT=1 para forzar."
    return 1
  fi

  log "${label}: waiting ${wait_seconds}s for expiration"
  sleep "$wait_seconds"
  return 0
}

cleanup_user_if_exists() {
  local id="$1"
  if [[ -z "$id" || "$id" == "null" ]]; then
    return
  fi

  request "DELETE" "/users/${id}" "" "" "" "0"
  if [[ "$RESPONSE_STATUS" != "204" && "$RESPONSE_STATUS" != "404" ]]; then
    fail "cleanup failed for user ${id}: http=${RESPONSE_STATUS} body=${RESPONSE_BODY}"
  fi
}

run_tests() {
  local suffix username email password user_id access_token new_access_token latest_access_token
  local access_exp_epoch refresh_exp_epoch old_refresh_token rotated_refresh_token
  suffix="$(date +%s)$RANDOM"
  username="auth_${suffix}"
  email="auth_${suffix}@example.com"
  password="StrongP@ss1"

  log "T1: POST /auth/register invalid content-type"
  request "POST" "/auth/register" '{"name":"A"}' "text/plain"
  assert_status "400" "T1"
  assert_jq '.success == false and .code == "INVALID_CONTENT_TYPE"' "T1"

  log "T2: POST /auth/register malformed JSON"
  request "POST" "/auth/register" '{"name":"broken",'
  assert_status "400" "T2"
  assert_jq '.success == false and .code == "MALFORMED_JSON"' "T2"

  log "T3: POST /auth/register invalid field type"
  request "POST" "/auth/register" "{\"name\":\"Auth\",\"lastName\":\"User\",\"username\":\"${username}_type\",\"email\":\"type_${email}\",\"password\":\"${password}\",\"avatar\":123}"
  assert_status "400" "T3"
  assert_jq '.success == false and .code == "INVALID_FIELD_TYPE"' "T3"

  log "T4: POST /auth/register unknown field"
  request "POST" "/auth/register" "{\"name\":\"Auth\",\"lastName\":\"User\",\"username\":\"${username}_unk\",\"email\":\"unk_${email}\",\"password\":\"${password}\",\"avatar\":\"\",\"unexpected\":true}"
  assert_status "400" "T4"
  assert_jq '.success == false and .code == "INVALID_PAYLOAD"' "T4"

  log "T5: POST /auth/register weak password"
  request "POST" "/auth/register" "{\"name\":\"Auth\",\"lastName\":\"User\",\"username\":\"${username}_weak\",\"email\":\"weak_${email}\",\"password\":\"weak\",\"avatar\":\"\"}"
  assert_status "400" "T5"
  assert_jq '.success == false and .code == "WEAK_PASSWORD"' "T5"

  log "T6: POST /auth/register valid"
  request "POST" "/auth/register" "{\"name\":\"Auth\",\"lastName\":\"User\",\"username\":\"${username}\",\"email\":\"${email}\",\"password\":\"${password}\",\"avatar\":\"\"}"
  assert_status "201" "T6"
  assert_jq '.success == true and (.data.id | length > 0)' "T6"
  user_id="$(jq -r '.data.id' <<<"$RESPONSE_BODY")"

  log "T7: POST /auth/register duplicate email"
  request "POST" "/auth/register" "{\"name\":\"Dup\",\"lastName\":\"Mail\",\"username\":\"${username}_dup\",\"email\":\"$(tr '[:lower:]' '[:upper:]' <<<"$email")\",\"password\":\"${password}\",\"avatar\":\"\"}"
  assert_status "409" "T7"
  assert_jq '.success == false and .code == "EMAIL_EXISTS"' "T7"

  log "T8: POST /auth/register duplicate username"
  request "POST" "/auth/register" "{\"name\":\"Dup\",\"lastName\":\"User\",\"username\":\"${username}\",\"email\":\"dup_${email}\",\"password\":\"${password}\",\"avatar\":\"\"}"
  assert_status "409" "T8"
  assert_jq '.success == false and .code == "USERNAME_EXISTS"' "T8"

  log "T9: POST /auth/login malformed JSON"
  request "POST" "/auth/login" '{"identity":"broken",'
  assert_status "400" "T9"
  assert_jq '.success == false and .code == "MALFORMED_JSON"' "T9"

  log "T10: POST /auth/login invalid credentials"
  request "POST" "/auth/login" "{\"identity\":\"${email}\",\"password\":\"WrongP@ss1\"}"
  assert_status "401" "T10"
  assert_jq '.success == false and .code == "INVALID_CREDENTIALS"' "T10"

  log "T11: POST /auth/login valid by username"
  request "POST" "/auth/login" "{\"identity\":\"${username}\",\"password\":\"${password}\"}"
  assert_status "200" "T11"
  assert_jq '.success == true and (.data.accessToken | length > 0)' "T11"

  log "T12: POST /auth/login valid by email"
  request "POST" "/auth/login" "{\"identity\":\"${email}\",\"password\":\"${password}\"}" "application/json" "" "1"
  assert_status "200" "T12"
  assert_jq '.success == true and (.data.accessToken | length > 0)' "T12"
  assert_cookie_header "$REFRESH_COOKIE_NAME" "T12"
  access_token="$(jq -r '.data.accessToken' <<<"$RESPONSE_BODY")"
  old_refresh_token="$(cookie_token_from_headers "$RESPONSE_HEADERS" "$REFRESH_COOKIE_NAME" || true)"
  [[ -n "$old_refresh_token" ]] || fail "T12: expected refresh cookie value"

  log "T13: GET /auth/me without token"
  request "GET" "/auth/me" "" "" "" "0"
  assert_status "401" "T13"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T13"

  log "T14: GET /auth/me with malformed bearer"
  request "GET" "/auth/me" "" "" "Bearer" "0"
  assert_status "401" "T14"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T14"

  log "T15: GET /auth/me with wrong scheme"
  request "GET" "/auth/me" "" "" "Token ${access_token}" "0"
  assert_status "401" "T15"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T15"

  log "T16: GET /auth/me with invalid token"
  request "GET" "/auth/me" "" "" "Bearer invalid.token.value" "0"
  assert_status "401" "T16"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T16"

  log "T17: GET /auth/me with token"
  request "GET" "/auth/me" "" "" "Bearer ${access_token}" "0"
  assert_status "200" "T17"
  jq -e --arg user_id "$user_id" '.success == true and .data.id == $user_id' >/dev/null 2>&1 <<<"$RESPONSE_BODY" || fail "T17: expected authenticated user ${user_id}. body=${RESPONSE_BODY}"

  log "T18: POST /auth/refresh without cookie"
  request "POST" "/auth/refresh" "" "" "" "0"
  assert_status "401" "T18"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T18"

  log "T19: POST /auth/refresh with cookie"
  request "POST" "/auth/refresh" "" "" "" "1"
  assert_status "200" "T19"
  assert_jq '.success == true and (.data.accessToken | length > 0)' "T19"
  new_access_token="$(jq -r '.data.accessToken' <<<"$RESPONSE_BODY")"
  [[ "$new_access_token" != "$access_token" ]] || fail "T19: expected a rotated access token"
  rotated_refresh_token="$(cookie_token_from_headers "$RESPONSE_HEADERS" "$REFRESH_COOKIE_NAME" || true)"
  [[ -n "$rotated_refresh_token" ]] || fail "T19: expected rotated refresh cookie"
  [[ "$rotated_refresh_token" != "$old_refresh_token" ]] || fail "T19: expected refresh token rotation"

  log "T20: POST /auth/refresh reuse old refresh token"
  request_with_cookie_value "POST" "/auth/refresh" "$REFRESH_COOKIE_NAME" "$old_refresh_token"
  assert_status "401" "T20"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T20"

  latest_access_token="$new_access_token"

  if [[ "$RUN_EXPIRY_TESTS" == "1" ]]; then
    log "T21: access token expiration behavior"
    access_exp_epoch="$(token_exp_epoch "$latest_access_token" || true)"
    if [[ -z "$access_exp_epoch" || "$access_exp_epoch" == "0" ]]; then
      fail "T21: could not parse access token exp"
    fi

    if wait_until_epoch_or_skip "T21 access token expiry" "$access_exp_epoch" "$MAX_WAIT_ACCESS_EXP_SECONDS"; then
      request "GET" "/auth/me" "" "" "Bearer ${latest_access_token}" "0"
      assert_status "401" "T21"
      assert_jq '.success == false and .code == "UNAUTHORIZED"' "T21"

      request "POST" "/auth/refresh" "" "" "" "1"
      assert_status "200" "T21-refresh"
      assert_jq '.success == true and (.data.accessToken | length > 0)' "T21-refresh"
      latest_access_token="$(jq -r '.data.accessToken' <<<"$RESPONSE_BODY")"
      [[ -n "$latest_access_token" ]] || fail "T21-refresh: expected new access token"
    fi

    log "T22: refresh token expiration behavior"
    request "POST" "/auth/login" "{\"identity\":\"${email}\",\"password\":\"${password}\"}" "application/json" "" "1"
    assert_status "200" "T22-login"
    refresh_exp_epoch="$(cookie_exp_epoch "$RESPONSE_HEADERS" "$REFRESH_COOKIE_NAME" || true)"
    if [[ -n "$refresh_exp_epoch" ]]; then
      if wait_until_epoch_or_skip "T22 refresh token expiry" "$refresh_exp_epoch" "$MAX_WAIT_REFRESH_EXP_SECONDS"; then
        request "POST" "/auth/refresh" "" "" "" "1"
        assert_status "401" "T22"
        assert_jq '.success == false and .code == "UNAUTHORIZED"' "T22"
      fi
    else
      log "T22: skipped (cookie expiry not available in headers)"
    fi
  fi

  log "T23: POST /auth/logout"
  request "POST" "/auth/logout" "" "" "" "1"
  assert_status "204" "T23"

  log "T24: POST /auth/logout without cookie (idempotent)"
  request "POST" "/auth/logout" "" "" "" "0"
  assert_status "204" "T24"

  log "T25: POST /auth/refresh after logout"
  request "POST" "/auth/refresh" "" "" "" "1"
  assert_status "401" "T25"
  assert_jq '.success == false and .code == "UNAUTHORIZED"' "T25"

  cleanup_user_if_exists "$user_id"
  log "All auth endpoint tests passed."
}

main() {
  require_cmd curl
  require_cmd jq

  COOKIE_JAR="$(mktemp)"
  trap 'rm -f "$COOKIE_JAR"' EXIT

  start_stack_if_needed
  wait_api
  run_tests
}

main "$@"
