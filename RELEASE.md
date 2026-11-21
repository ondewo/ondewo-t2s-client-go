# Release History

*****************

## Release ONDEWO T2S Go Client 6.7.1

### New Features

* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) Offline-token start-up, as the python SDK has it: a handed-off Keycloak OFFLINE refresh token (`RefreshToken`) replaces the password. The provider then starts with ONE `grant_type=refresh_token` exchange and never sends a password grant, not even when the token is refused, because Keycloak's brute-force detection counts failed password logins and N processes sharing one technical user would lock it. A Keycloak config needs `password` or `RefreshToken`; with both, the refresh token wins.
* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) `client.Config` takes `KeycloakURL`, `Realm`, `ClientID`, `Username`, `Password`, `RefreshToken`, `TokenExpirationInS` and `KeycloakInsecureSkipVerify`, and `client.NewChannel` attaches `authorization: Bearer <token>` from the shared `auth.KeycloakTokenProvider` to every RPC; a refused start-up returns `*auth.KeycloakAuthenticationError`, an expired token that cannot be refreshed fails the RPC with `Unauthenticated`.

### Bug Fixes

* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) Errors are named by their bare type in logs and error messages, as the python SDK names them.

### Improvements

* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) A failed token refresh never ends the refreshing: every failure (an unreachable or timed-out token endpoint, a non-2xx answer, a body that is not a JSON object, a missing `access_token`, a non-numeric `expires_in`) is retried with a full-jitter backoff from 5 s doubling up to 300 s, which the background refresh and the per-call refresh share, so a Keycloak outage costs one token request per backoff step instead of one per call. A success resets the backoff.
* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) A token that is still valid keeps being served while its refresh fails; only an EXPIRED token that cannot be refreshed fails, with a Keycloak authentication error that chains the underlying cause. A malformed token response changes no provider state, and an error about a response without an `access_token` names its field names only, never their values.
* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) One provider is shared per credential set: the registry key is a SHA-256 over the Keycloak connection and credential fields (the password and the refresh token enter it only as hashes), so two configs of different tenants never share a provider and all clients of one identity log in once.
* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) The token request times out after 10 s, quoted Keycloak error bodies are cut at 500 characters, each failed attempt logs one WARNING (attempt number, retry delay, remaining token validity) and the first success afterwards one INFO; no error message and no log line carries a password or a token.
* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) The behaviour is pinned by the same test catalogue as the python SDK (`test_keycloak_refresh_resilience.py`, `test_client_config_refresh_token.py`), plus edge cases: every failure kind, the backoff ladder and its bounds, concurrent callers during an outage (one token request), malformed responses, credential-free logs and the credential-keyed registry.

*****************

## Release ONDEWO T2S Go Client 6.7.0

### Improvements

* Tracking API Version [6.7.0](https://github.com/ondewo/ondewo-t2s-api/releases/tag/6.7.0) ( [Documentation](https://ondewo.github.io/ondewo-t2s-api/) )
* [[OND236-120]](https://ondewo.atlassian.net/browse/OND236-120) Generated with ondewo-proto-compiler 5.17.0 (protoc 35.1). New runtime floors: `google.golang.org/grpc` v1.85.0 (carries the CVE-2026-84445 fix), `google.golang.org/protobuf` v1.36.12, genproto v0.0.0-20261116203732-fd2c7689a808; `go.mod` declares `go 1.26.0`.

*****************

## Release ONDEWO T2S Go Client 6.6.1

### New Features

* `client.NewChannel(cfg client.Config, opts ...grpc.DialOption)` opens the gRPC connection for
  plaintext, TLS or mutual TLS, following the TLS contract of the python SDKs
  (`ondewo-client-utils` 4.1.1), so one set of certificates works with every ONDEWO client.
  `client.Config` carries `Host`, `Port`, `Insecure`, an optional `Logger` and the PEM **content**
  (never a file path) of the CA (`GrpcCert`, empty = system roots) and of an optional client
  identity (`GrpcClientCert` / `GrpcClientKey`).
* Refused with an error before gRPC sees them: half a client identity, a certificate and key that
  do not form a pair, a `GrpcCert` that holds no PEM certificate (typically a path), and
  `Insecure: true` combined with a client identity. No error message contains a PEM, a key or the
  whole `Config`; they name the field and `host:port`.
* A plaintext connection logs a warning naming `host:port` through `log/slog` (`Config.Logger`, or
  `slog.Default()`); the package never configures logging. Every `fmt` verb and `slog` rendering of
  a `Config` redacts the client key.
* Bare IPv6 literal hosts are bracketed (`::1` -> `[::1]:<port>`); bracketed hosts and hosts with a
  scheme are used as they are. PEMs with CRLF line endings work.
* Connection defaults of the python SDKs that grpc-go exposes: keepalive pings every 30 s with a
  20 s timeout, only while an RPC is active; 5 s maximum reconnect backoff; 2^31-1 byte maximum
  message size in both directions. Every `grpc.DialOption` passed in is applied after them and wins.
  Documented gaps: grpc-go has no `http2.max_pings_without_data` and only one keepalive timeout, and
  no per-method retry policy is configured (only gRPC's transparent retries apply).

### Improvements

* `tests/tls_test.go` runs real TLS and mutual-TLS handshakes against an in-process server with a
  PKI generated per run; the 100% coverage gate now spans `auth/` and `client/`.
* `tests/release_notes_test.go` pins the `RELEASE.md` slice the GitHub release body is built from:
  the Makefile's perl range, the spelling of every release heading, the `*****` separator closing
  every section, and non-empty notes for the current version.
* README: new section "TLS, mutual TLS and certificates" (modes, connection defaults, a test PKI with
  openssl, TLS security notes, troubleshooting).
* The ONDEWO proto compiler submodule is pinned to 5.15.2 (was 5.15.1).

*****************

## Release ONDEWO T2S Go Client 6.6.0

### New Features

* Initial release of the ONDEWO T2S (Text-to-Speech) gRPC client for Go. The module
  ships the stubs generated from the [ONDEWO T2S API](https://github.com/ondewo/ondewo-t2s-api)
  by version 5.15.1 of the
  [ONDEWO Proto Compiler](https://github.com/ondewo/ondewo-proto-compiler): one `*.pb.go` of
  messages and one `*_grpc.pb.go` of service stubs per `.proto`, below `api/ondewo/t2s/`,
  compiled against the `google.golang.org/protobuf` and `google.golang.org/grpc` runtimes pinned by
  the compiler image.
* `make build` reproduces the whole client from the two submodules — proto compiler image, stub
  generation and `go build` — and `make check_build` asserts that every `.proto` of the API
  produced a stub.
* A go test suite under `tests/`, run by `.github/workflows/ci.yml` on go 1.25 and 1.27 against the
  committed stubs. It needs no network and no ONDEWO server: gRPC calls travel over an in-memory
  `bufconn` listener, so the generated client stub, the generated server stub and a real HTTP/2
  connection are all exercised in-process. Messages round-trip through the wire format, every
  generated unary stub is called, and the two generators' views of the service (`grpc.ServiceDesc`
  and the proto descriptor) are cross-checked against each other.

### Improvements

* Version numbering follows the fleet rule: the client version matches the ONDEWO T2S API in
  major and minor version, so the first release is `6.6.0` rather than a `0.x`. The heading
  above and `ONDEWO_T2S_VERSION` in the `Makefile` now agree on it — they did not before, and
  `make build_gh_release` slices its release body out of this file by exactly that string, so the
  mismatch would have published a GitHub release with an empty body.
* `make test_coverage` gates the build on the coverage of the HAND-WRITTEN code (`auth/`, currently
  100%) and fails when it ends up measuring nothing at all; `make test_coverage_generated` reports
  the generated stubs' figure (15.5%) without ever gating on it. `make check_stubs` runs
  first in CI and fails when `api/` is absent or empty, so the pipeline cannot go green by finding
  no work to do.
* The release targets tag each release twice on the same commit: with the ONDEWO release number
  (`6.6.0`) that the rest of the fleet uses, and with the `v`-prefixed spelling (`v6.6.0`) that is
  the only tag shape the Go module resolver accepts. `make publish_go_module` then warms
  `proxy.golang.org` so the new version is immediately installable with `go get`.

*****************
