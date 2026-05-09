"""
Test that coi shell with [tool] name = "pi" starts the real pi binary
and injects settings into ~/.pi/agent/settings.json.

Verifies that:
1. Writing [tool] name = "pi" to .coi/config.toml is accepted
2. coi shell starts the container and launches the real pi binary
3. Pi's startup UI appears on screen
4. ~/.pi/agent/settings.json is created in the container with injected settings

No API key is required - pi displays a provider selection screen without one.
"""

import json
import os
import subprocess
import time

from pexpect import EOF, TIMEOUT

from support.helpers import (
    calculate_container_name,
    spawn_coi,
    wait_for_container_ready,
    wait_for_text_on_screen,
)


def test_pi_tool_starts_session(coi_binary, cleanup_containers, workspace_dir):
    """
    Smoke test: coi shell with tool = "pi" launches the real pi binary
    and injects the sandbox config into ~/.pi/agent/settings.json.

    Flow:
    1. Write .coi/config.toml with [tool] name = "pi" to the workspace
    2. Start coi shell (no COI_USE_DUMMY - use the real binary)
    3. Wait for container setup to complete
    4. Wait for pi's startup UI to appear on screen
    5. While TUI is running, use incus exec to inspect ~/.pi/agent/settings.json
    6. Ctrl+C to exit the TUI, then poweroff
    """
    config_dir = os.path.join(workspace_dir, ".coi")
    os.makedirs(config_dir, exist_ok=True)
    config_path = os.path.join(config_dir, "config.toml")
    with open(config_path, "w") as f:
        f.write('[tool]\nname = "pi"\n')

    container_name = calculate_container_name(workspace_dir, 1)

    child = spawn_coi(
        coi_binary,
        ["shell"],
        cwd=workspace_dir,
        timeout=120,
    )

    wait_for_container_ready(child, timeout=60)

    # Wait for pi's startup UI
    pi_started = False
    try:
        wait_for_text_on_screen(child, "pi", timeout=60)
        pi_started = True
    except Exception:
        pass

    # While TUI is running, inspect ~/.pi/agent/settings.json via incus exec
    settings_exists = False
    try:
        result = subprocess.run(
            [
                "incus",
                "exec",
                container_name,
                "--",
                "python3",
                "-c",
                (
                    "import json; "
                    "d = json.load(open('/home/code/.pi/agent/settings.json')); "
                    "print(json.dumps(d))"
                ),
            ],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode == 0:
            settings_exists = True
    except Exception:
        pass

    # Stop pi with Ctrl+C, fall back to bash, then poweroff
    child.send("\x03")
    time.sleep(1)
    child.send("\x03")
    time.sleep(2)

    child.send("sudo poweroff")
    time.sleep(0.3)
    child.send("\x0d")

    try:
        child.expect(EOF, timeout=60)
    except TIMEOUT:
        pass

    try:
        child.close(force=False)
    except Exception:
        child.close(force=True)

    time.sleep(5)

    subprocess.run(
        [coi_binary, "container", "delete", container_name, "--force"],
        capture_output=True,
        timeout=30,
    )

    assert pi_started, (
        "coi shell with [tool] name = 'pi' should launch the pi binary "
        "and display its startup UI"
    )
    assert settings_exists, (
        "~/.pi/agent/settings.json should exist with injected settings"
    )
