# Local QA: settings and appearance

## Repeatable checks

1. Windows x64, interactive desktop, Go in PATH: run `build.cmd`.
2. Copy `vpn-monitor.exe` to an empty writable directory. Never run settings QA
   against a user installation or an existing configuration.
3. Run `python settings_verify.py <isolated-folder>/vpn-monitor.exe` (Python + Pillow).
   It refuses existing JSON, uses actual ICMP loopback and a local UDP DNS fixture,
   writes evidence under `<isolated-folder>/qa`, and removes its test JSON.
   `compact_verify.py` is a compatibility entry point for this same deterministic suite;
   do not run the two concurrently. The retired external-network main-Apply test
   is no longer applicable because timeout and pause now live in settings.
4. Run `python icon_verify.py <isolated-folder>/vpn-monitor.exe` for embedded PE icon
   resources, WM_GETICON/class handles and shell/runtime icon pixel identity.
5. Run `python mouse_tooltip_verify.py <isolated-folder>/vpn-monitor.exe` for actual
   mouse pin toggles and live tooltips. Local logs/screenshots must not be published.

## Acceptance

- Main is 380×310 logical, journal collapsed, no old 101/102/107 controls or empty row.
  Journal expands to 380×494 without moving the top-left; header controls do not overlap.
- Settings uses native bordered tab-stop edits and a native 0–70 trackbar plus percent.
  Keyboard movement updates its percent. Settings itself never has WS_EX_LAYERED.
- Transparency is applied only after a successful Save: 0% → alpha255,
  40% → alpha153, 70% → alpha77, verified by GetLayeredWindowAttributes.
  Cancel and failed saves leave config, disk and actual main alpha unchanged.
- Pin modifies actual WS_EX_TOPMOST; saving opacity while pinned preserves it,
  and unpinning preserves alpha. No click-through style is added.
- Pause/Continue in settings acts immediately and survives Cancel, freezes probes,
  cycles and counters after settling, and is not persisted between launches.
- Appearance-only changes preserve node results, chart samples and outage counter.
- Real four-ICMP loopback + fixture.test DNS succeeds; saved timeout appears in all rows.
  Timeout validation covers invalid strings and 1/60000 boundaries.
- Restart loads targets, timeout and transparency. Old version1 without the optional
  percent is opaque. Invalid percent and malformed JSON are preserved with visible errors.
- Defaults fills timeout2000 and transparency0 without applying until Save.
  Changing fourth identity while paused still resets graph/counter/result.
- GUI subsystem/no console, zero-exit close and unchanged embedded icons are checked.
- Uncached Go tests cover persistence/backward compatibility, validation, alpha rounding,
  preservation of existing extended-style bits, history retention and smaller layout,
  in addition to prior real networking/episode/generation tests. Run `go vet` too.

## Observed local verification

Actual Windows desktop: 120 DPI (125%), main outer size 475×388 pixels.
Settings, opaque/translucent main and expanded-journal screenshots were captured
and visually inspected; fields, slider, percent, pause explanation and Save/Cancel fit.
The test config is removed after QA. The ICO and `.syso` resource remain unchanged.

Not claimed: remote CI/release verification, screen-reader conformance, mixed-DPI
monitor migration, crash/power-loss durability or simultaneous multi-instance edits.
The executable remains unsigned. No commits, push, tags or release writes are made here.
