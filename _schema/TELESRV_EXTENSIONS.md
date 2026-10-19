# telesrv protocol extensions

Rules for adding RPCs that exist only between the `telesrv` server and the
project's own clients (Telegram Desktop, Android, iOS forks). They are not an
official Telegram Layer and are never sent to official servers.

## Where the definitions live

- **One source of truth:** `_schema/telesrv.tl`, registered as an entry in
  `manifest.json` → `overlays` (the same mechanism as `legacy.tl`). Overlays are
  merged into every exact profile (225 and up) and generate ordinary
  `tg.*Request` types and `tlprofile.SemanticMethod*` IDs.
- **Not** `client_rpc_overlays`: that mechanism only adapts a client-private
  serialization onto an *existing* canonical method by semantic name, and
  aborts generation for a method with no canonical counterpart.
- Clients copy or mirror this file; they never become the source:
  - Telegram Desktop: copy `telesrv.tl` byte for byte into
    `Telegram/SourceFiles/mtproto/scheme/` and add it to `scheme_files` in
    `Telegram/cmake/td_scheme.cmake`. Never append to `api.tl`; it is replaced on
    every upstream Layer update, and a Layer import from that checkout would pick
    the extension up as if it were official.
  - iOS: hand-written serializers in a separate file (for example
    `submodules/TelegramApi/Sources/ApiTelesrv.swift`), attached with
    `extension Api.functions { public enum telesrv {} }`. Do not edit the
    generated `Api*.swift` files.
  - Android: hand-written classes kept apart from the upstream `TLRPC.java`
    where possible.
  - Every hand-written serializer names the `telesrv.tl` line and wire ID it
    implements in a comment.

## Rules

1. **Own namespace only.** Every definition is `telesrv.<name>`. Never add to
   `auth.`, `account.` or any official namespace: an official definition with
   the same name in a later Layer would collide in `mergeSchema`, and a separate
   namespace keeps each client's code in its own file.
2. **Reuse official types; avoid new constructors.** Parameters and results
   should be existing types (`account.Password`, `auth.Authorization`,
   `InputCheckPasswordSRP`, `DataJSON`, ...). A new method costs each mobile
   client one request serializer. A new type costs a hand-written deserializer
   plus a registration in each client's generated parser table
   (`Api0.swift` `parsers`, the Android constructor switch), which conflicts on
   every upstream Layer update. A PR that adds a constructor states why no
   official type fits.
3. **Start every method with `flags:#`**, even when no optional field exists
   yet. TL can only add fields behind an existing flags word; without one, any
   later change needs a new wire ID on every platform.
4. **Wire IDs are explicit and immutable.** `validateExplicitIDs` requires an
   explicit `#id`. Use the TL CRC32 of the declaration, and check it against
   every loaded profile for collisions. Once released to any client, an ID and
   its field layout never change; a semantic change is a new method, and the
   old one stays until no supported client sends it.
5. **Capability discovery through `help.getAppConfig`.** The server advertises
   each feature in the `telesrv_features` string array (for example
   `"password_login"`). Clients show an entry point and call the methods only
   when the feature is listed, so they degrade cleanly against official servers
   and older `telesrv` builds.
6. **Data-only features may use a generic envelope.** A feature that only
   exchanges plain data (settings, toggles, lists without TL objects) can go
   through one `telesrv.invoke flags:# method:string params:DataJSON = DataJSON`
   method instead of a new TL definition, so clients need no further TL work.
   Anything that must return TL objects (`User`, `Updates`,
   `auth.Authorization`, ...) gets a typed method.
7. **Defer a client generator.** While the extension stays small, hand-written
   mobile serializers are cheaper than tooling. Once custom constructors grow
   past roughly a dozen, add a gotdgen mode that emits Swift/Java serializers
   from `telesrv.tl` only.

## Adding or changing a definition

1. Edit `_schema/telesrv.tl` and commit it on its own.
2. Register or update the `overlays` entry in `manifest.json`: `repository`
   is this fork, `path` is `_schema/telesrv.tl`, `commit`/`blob` come from the
   commit in step 1, and `sha256` is the file's digest. Offline generation
   verifies only `sha256`; the import/sync workflow verifies the Git
   provenance.
3. Regenerate and run the gates from `layers/README.md` ("Adding a Layer",
   step 4), then validate `telesrv` through a temporary `go.work` before
   tagging a release.
4. Server (`telesrv`): register the handler with `registerRPC`, add
   pre-authorization methods to `rpcAllowedWithoutAuthorization` in
   `internal/rpc/auth_gate.go`, and advertise the feature in
   `telesrv_features`.
5. Clients: mirror the definition per "Where the definitions live" above.
