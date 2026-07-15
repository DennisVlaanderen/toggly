# FQDP — Feature Query & Definition Protocol

Version: 0.1

## Overview

FQDP is a lightweight, language-agnostic protocol designed for high-throughput feature flag queries and subscriptions. It operates at OSI Layer 4 over raw TCP sockets rather than relying on HTTP or WebSocket semantics. This document specifies a binary-first wire format with an optional JSON/text fallback and a token-based handshake for authentication.

Goals

- Low-latency queries and push updates for large numbers of clients
- Simple subscription semantics with initial snapshot + deltas
- Language-agnostic: binary framing primary, JSON alternative for ease of debugging
- Secure connections via TLS and token-based handshake

Transport

- Raw TCP socket at OSI Layer 4 (recommended), TLS optional (STARTTLS-style or direct TLS)
- Server accepts connections and keeps a lightweight session per client
- No HTTP upgrade or WebSocket framing layer; framing is provided by the protocol itself
- Idle/keepalive heartbeats supported to detect dead peers

Framing

- Primary framing: 4-byte big-endian length prefix followed by binary payload
  - `[u32 length][payload bytes]`
- Payload is a message encoded as binary (recommended) containing a message-type byte + typed fields. For clarity and interop, a JSON text variant using newline-delimited JSON is described as a fallback.

Message Model

- Each message begins with a 1-byte MessageType tag. Additional fields follow per message type.
- MessageType:
  - 0x01: Handshake (client -> server)
  - 0x02: HandshakeAck (server -> client)
  - 0x10: Query (client -> server)
  - 0x11: QueryResponse (server -> client)
  - 0x20: Subscribe (client -> server)
  - 0x21: Unsubscribe (client -> server)
  - 0x22: Update (server -> client)
  - 0x30: Ack (both directions)
  - 0x31: Nack (server -> client)
  - 0x40: Heartbeat (both)
  - 0x50: Error (server -> client)
  - 0xF0: Admin (server/client, restricted)

Handshake & Authentication

- Client -> Server: Handshake message (0x01)
  - Fields (binary framing):
    - 1 byte MessageType (0x01)
    - 1 byte protocol version major
    - 1 byte protocol version minor
    - 2 bytes reserved
    - 2 byte token length (u16 BE)
    - token bytes (UTF-8)
    - optional client id length + client id
  - Semantics: Client sends token; server validates token, creates session, and replies with HandshakeAck.

- Server -> Client: HandshakeAck (0x02)
  - Fields:
    - 1 byte MessageType (0x02)
    - 1 byte accepted protocol major
    - 1 byte accepted protocol minor
    - 16 byte session id (opaque bytes, e.g., UUID)
    - 1 byte server flags
    - optional welcome message length + bytes
  - On failure, server may send Error (0x50) and close connection.

Query / Response

- Query (0x10): client requests the current value of a flag or set of flags.
  - Fields: MessageType, query id (u32), flags payload (e.g., one or more names, length-prefixed)
- QueryResponse (0x11): server replies with flag values and a version vector or sequence number for each item.

Subscribe / Update

- Subscribe (0x20): client requests subscription to one or more flags or a prefix/namespace.
  - Fields: MessageType, subscription id (u32), criteria (length-pref list)
- Unsubscribe (0x21): cancel subscription id
- Update (0x22): server pushes changes matching subscription criteria.
  - Fields: MessageType, subscription id (u32), sequence number (u64), payload (flag key + value + metadata + version)

Acks & Ordering

- Ack (0x30): acknowledges receipt of server update(s). Contains message id or sequence cursor.
- Server may require explicit acks for critical Admin messages; normal updates can be applied best-effort with client-side ordering via sequence numbers.

Heartbeat & Keepalive

- Heartbeat (0x40): empty payload messages to keep NAT/firewalls alive and detect dead peers. Interval configured per deployment.

Errors

- Error (0x50): error code (u16) + message. On auth failure, server should close the connection after sending an Error.

Admin

- Admin (0xF0): restricted commands (e.g., force refresh, server diagnostics). Access controlled by token/ACL.

Versioning & Compatibility

- Use major.minor version numbers in the handshake. If major differs, server should reject. Minor differences may be tolerated with feature negotiation bits in HandshakeAck.

Reconnect and Backfill

- On reconnect, clients perform Handshake. After successful handshake, client may send last known sequence numbers per subscription to request deltas/backfill.
- Server responds with missing updates in order and then transitions to push updates.

Security Considerations

- Tokens should be short-lived or scoped. Recommend bearer tokens (opaque) or JWT carrying scope.
- TLS is strongly recommended in production.
- Rate-limit unauthenticated or new connections per IP and apply request rate limits per session.

Binary vs JSON examples

- Binary (length-prefixed) example (pseudocode):
  - `[u32 length][0x01][0x01][0x00][0x00][u16 token_len][token bytes]`
- JSON newline-delimited handshake example (fallback):
  - {
    "type":"handshake",
    "version":"1.0",
    "token":"..."
    }\n

Operational Notes

- Servers should persist flag changes to a durable store and replicate changes to other nodes. The protocol focuses on efficiently delivering deltas and snapshots to clients.
- Prefer idempotent updates and include version/sequence numbers so clients can safely apply or request backfill.

Appendix: wire-format reference (summary)

- `[u32 length][payload bytes]`
- payload: `[u8 MessageType][body...]`

## Binary Field Layouts

The following specifies exact binary layouts (big-endian) for commonly used messages. Sizes shown are bytes.

1) Handshake (client -> server) — MessageType 0x01

   - [u32 length]                : total payload length (not including this length field)
   - [u8 0x01]                   : MessageType
   - [u8 ver_major]              : protocol major version
   - [u8 ver_minor]              : protocol minor version
   - [u16 reserved]              : reserved for flags (set 0)
   - [u16 token_len]             : token length in bytes (N)
   - [N bytes token]             : token UTF-8
   - [u16 client_id_len]         : client id length (M, optional, 0 if absent)
   - [M bytes client_id]         : client id UTF-8 (optional)

2) HandshakeAck (server -> client) — MessageType 0x02

   - [u32 length]
   - [u8 0x02]
   - [u8 ver_major]
   - [u8 ver_minor]
   - [16 bytes session_id]
   - [u8 server_flags]
   - [u16 welcome_len]
   - [welcome_len bytes welcome_msg]

3) Query (client -> server) — MessageType 0x10

   - [u32 length]
   - [u8 0x10]
   - [u32 query_id]
   - [u16 key_count]
   - for each key:
     - [u16 key_len]
     - [key_len bytes key]

4) QueryResponse (server -> client) — MessageType 0x11
    - [u32 length]
    - [u8 0x11]
    - [u32 query_id]
    - [u16 item_count]
    - repeated items:
      - `[u16 key_len][key]`
      - [u8 value_type] (0=bool,1=int,2=string,3=json)
      - `[u32 value_len][value_bytes]`
      - [u64 version]

5) Subscribe (client -> server) — MessageType 0x20
    - [u32 length]
    - [u8 0x20]
    - [u32 subscription_id]
    - [u16 criteria_count]
    - repeated criteria: `[u16 len][bytes]`

6) Update (server -> client) — MessageType 0x22
    - [u32 length]
    - [u8 0x22]
    - [u32 subscription_id]
    - [u64 sequence]
    - [u16 change_count]
    - repeated changes:
      - `[u16 key_len][key]`
      - [u8 value_type]
      - `[u32 value_len][value_bytes]`
      - [u64 version]

7) Ack (both directions) — MessageType 0x30
    - [u32 length]
    - [u8 0x30]
    - [u32 ref_id] (query_id or sequence low bits)

8) Heartbeat (both) — MessageType 0x40
    - [u32 length]
    - [u8 0x40]

9) Error (server -> client) — MessageType 0x50
     - [u32 length]
     - [u8 0x50]
     - [u16 code]
     - [u16 msg_len]
     - [msg_len bytes message]

## Hex Frame Examples

Example A — Handshake (binary length-prefixed) with token "tok-ABC":

- Fields breakdown:
  - MessageType: 0x01
  - ver_major: 0x01, ver_minor: 0x00
  - reserved: 0x0000
  - token_len: 0x0007 (7 bytes)
  - token bytes: 74 6f 6b 2d 41 42 43 ("tok-ABC")
  - client_id_len: 0x0000

- Payload bytes (MessageType+body):
  01 01 00 00 00 00 07 74 6f 6b 2d 41 42 43 00 00

- Add 4-byte length prefix (payload length = 1+1+1+2+2+7+2 = 16 -> 0x00000010):
  00 00 00 10 01 01 00 00 00 00 07 74 6f 6b 2d 41 42 43 00 00

Hex (grouped):
  00000010 01 01 00 00 0007 746f6b2d414243 0000

Example B — HandshakeAck (server accepts), session id shown as 16 bytes 00..0f:

- Payload (no welcome message):
  02 01 00 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 00

- 4-byte length prefix = payload length (1 +1 +1 +16 +1 +2 +0 = 22 -> 0x16):
  00 00 00 16 02 01 00 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 00

Example C — Subscribe to key "feature.alpha":

- key bytes: 66 65 61 74 75 72 65 2e 61 6c 70 68 61 (13 bytes)
- subscription_id: 0x00000001
- criteria_count: 0x0001

- Payload:
  20 00 00 00 01 00 01 00 0d 66 65 61 74 75 72 65 2e 61 6c 70 68 61

- length prefix (payload len = 1 +4 +2 + (2+13) = 22 -> 0x16):
  00 00 00 16 20 00 00 00 01 00 01 00 0d 66 65 61 74 75 72 65 2e 61 6c 70 68 61

### Notes

- The hex examples above are illustrative; actual implementations should produce consistent big-endian integer encoding and must handle partial reads/writes on TCP streams.
- For debugging, the newline-delimited JSON form can be used instead of binary; production clients should prefer binary framing for performance.

### Comprehensive Hex Examples (all message types)

Below are worked hex examples for each message type. All multi-byte integers are big-endian.

1) Handshake (client -> server) — token "tok-ABC" (example A)

     - Already shown above (payload length 0x00000010):
       00 00 00 10 01 01 00 00 00 00 07 74 6f 6b 2d 41 42 43 00 00

2) HandshakeAck (server -> client) — accepted (example B)

     - Already shown above (payload length 0x16):
       00 00 00 16 02 01 00 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 00

3) Query (client -> server) — query_id=0x0000002a, keys ["feature.alpha","feature.beta"]

     - key1 bytes (13): 66 65 61 74 75 72 65 2e 61 6c 70 68 61
     - key2 bytes (12): 66 65 61 74 75 72 65 2e 62 65 74 61
     - payload (MessageType 0x10):
       10 00 00 00 2a 00 02 00 0d 66 65 61 74 75 72 65 2e 61 6c 70 68 61 00 0c 66 65 61 74 75 72 65 2e 62 65 74 61
     - length prefix: payload len = 1 +4 +2 + (2+13) + (2+12) = 40 -> 0x00000028
       00 00 00 28 10 00 00 00 2a 00 02 00 0d 66 65 61 74 75 72 65 2e 61 6c 70 68 61 00 0c 66 65 61 74 75 72 65 2e 62 65 74 61

4) QueryResponse (server -> client) — query_id=0x0000002a, two items

     - item1 key="feature.alpha" (13), value_type=0 (bool), value_len=1 (0x01), value=0x01 (true), version=0x0000000000000005
     - item2 key="feature.beta" (12), value_type=2 (string), value_len=0x00000005, value="hello", version=0x0000000000000006
     - payload (MessageType 0x11):
       11 00 00 00 2a 00 02 00 0d 66 65 61 74 75 72 65 2e 61 6c 70 68 61 00 01 00 00 00 01 00 00 00 00 00 00 00 05 00 0c 66 65 61 74 75 72 65 2e 62 65 74 61 02 00 00 00 05 68 65 6c 6c 6f 00 00 00 00 00 00 00 06
     - length prefix: compute total payload length accordingly and prefix.

5) Subscribe (client -> server) — subscription_id=0x00000001, criteria_count=1, criteria[0]="feature.alpha"

     - payload (MessageType 0x20) shown earlier (example C). Note: each criterion is length-prefixed (`[u16 len][bytes]`) so multiple criteria are unambiguous.

6) Update (server -> client) — subscription_id=0x00000001, sequence=0x000000000000000a, change_count=1, change: key="feature.alpha", value_type=0 (bool), value=0x00 (false), version=0x0000000000000006

     - key bytes (13): 66 65 61 74 75 72 65 2e 61 6c 70 68 61
     - payload (MessageType 0x22):
       22 00 00 00 01 00 00 00 00 00 00 00 0a 00 01 00 0d 66 65 61 74 75 72 65 2e 61 6c 70 68 61 00 00 00 00 00 00 00 06
     - length prefix added as usual.

7) Ack (both) — ref_id=0x0000002a

     - payload: 30 00 00 00 2a
     - length prefix = 0x00000005 -> 00 00 00 05 30 00 00 00 2a

8) Heartbeat (both)

     - payload: 40
     - length prefix = 0x00000001 -> 00 00 00 01 40

9) Error (server -> client) — code=0x0001, message="auth failed"

      - message bytes (11): 61 75 74 68 20 66 61 69 6c 65 64
      - payload: 50 00 01 00 0b 61 75 74 68 20 66 61 69 6c 65 64
      - length prefix accordingly.

10) Admin (example) — Admin force-refresh command with body "refresh_all"

      - message bytes (11): 72 65 66 72 65 73 68 5f 61 6c 6c
      - payload: f0 00 00 00 00 0b 72 65 66 72 65 73 68 5f 61 6c 6c
