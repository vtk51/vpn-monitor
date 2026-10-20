# Local QA: persistent settings

## Repeatable checks

1. Windows x64, interactive desktop, Go in PATH: run `build.cmd`.
2. Copy `vpn-monitor.exe` into a new empty, writable folder. Never run settings QA
   against a user installation or an existing configuration.
3. Run `python settings_verify.py <isolated-folder>/vpn-monitor.exe` (Python + Pillow).
   It refuses an existing JSON file, uses real ICMP loopback and an actual local UDP
   DNS fixture, writes evidence under `<isolated-folder>/qa`, restores defaults through
   the UI and removes the test JSON. An empty/appended diagnostic log may remain.
4. Optional: `python icon_verify.py <isolated-folder>/vpn-monitor.exe` checks actual
   embedded PE icon resources and window/shell icon pixel identity. This older smoke
   changes/saves timeout; remove the isolated test JSON afterwards.
5. `compact_verify.py` remains an extended external-network test. It needs actual
   fourth-target failure at 1 ms and subsequent recovery, so it is not a deterministic
   gate. Its timeout changes now apply to the next coherent cycle, not each probe.

## Settings QA acceptance

- Compact main dimensions remain 380×353 logical, journal collapsed initially.
- Gear/pin/minimize/close hit rectangles do not overlap; settings is a real owned
  native window, with tab-stop native edit controls and bordered fields, no scrollbars.
- Cancel, defaults + Cancel and validation errors leave live configuration and disk
  untouched. Save failure at the actual target path leaves prior live settings intact.
- Four distinct names with 127.0.0.1 run actual successful ICMP. Custom `fixture.test`
  is observed on the exact configured local DNS UDP endpoint, no production fake data.
- Save updates visible labels, addresses, rows, chart target, and persisted JSON.
- Main timeout Apply persists to the same file; invalid input preserves the old value.
- Restart loads saved names/addresses/timeout. Malformed original JSON remains unchanged
  until explicit Save; defaults and an error are visible instead of a crash.
- Changing fourth identity while paused clears samples, counter and prior result.
- Pin changes actual WS_EX_TOPMOST; pause settles and freezes probes/cycles; journal,
  verbosity, resume, no-console GUI subsystem, zero-exit close are exercised.
- Go tests reject stale-generation publications, verify save failures, round-trip
  persistence and validation; existing episode/history/pause tests remain enabled.

## Observed local verification

Verified on the actual Windows desktop at 120 DPI (125%): main outer rectangle
475×441 pixels, all local settings GUI assertions passed. Screenshots of settings,
invalid input, live loopback main and expanded journal were captured and visually
inspected. Initial QA caught unwanted EDIT horizontal scrollbars; the final native
fields use explicit client-edge borders and tab stops, and the regression checks
actual control styles. The embedded ICO and `.syso` resource are unchanged.

Not claimed: remote CI/release verification, screen-reader conformance, mixed-DPI
monitor migration, crash/power-loss durability or simultaneous multi-instance edits.
No commit, push, tag, or release is performed by this task. Local configs/logs/screenshots
must not be included in a source commit or release asset.
