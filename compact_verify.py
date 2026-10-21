"""Compatibility entry point for compact/settings/appearance GUI QA.
Only run against an isolated executable copy (no existing JSON).
"""
from pathlib import Path
import runpy
runpy.run_path(str(Path(__file__).with_name("settings_verify.py")), run_name="__main__")
