"""
Test auto-context injection for pi.

Verifies that:
1. When auto_context is enabled (default) and tool is pi, the
   AGENTS.md file is created in ~/.pi/agent/ with the sandbox context.
2. settings.json remains valid (pi has no permission bypass system).
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
)


def test_auto_context_pi_instructions_injected(coi_binary, cleanup_containers, workspace_dir):
    """
    Test that pi's auto-context file is created at ~/.pi/agent/AGENTS.md
    when auto_context is enabled.

    Flow:
    1. Write .coi/config.toml with [tool] name = "pi"
    2. Start coi shell
    3. While running, inspect ~/.pi/agent/settings.json via incus exec
    4. Verify it exists and is valid (empty or with user settings)
    5. Verify ~/.pi/agent/AGENTS.md exists (auto-context file for pi)
    6. Verify AGENTS.md contains COI Sandbox Context
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

    # Give pi a moment to start
    time.sleep(5)

    # Check if ~/.pi/agent/settings.json exists and is valid
    settings_valid = False
    settings_content = ""
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
            settings_content = result.stdout.strip()
            # Verify it's valid JSON (even if empty object)
            json.loads(settings_content)
            settings_valid = True
    except Exception:
        pass

    # Check if ~/.pi/agent/AGENTS.md exists and contains COI context
    agents_md_exists = False
    agents_md_has_context = False
    try:
        result = subprocess.run(
            [
                "incus",
                "exec",
                container_name,
                "--",
                "test",
                "-f",
                "/home/code/.pi/agent/AGENTS.md",
            ],
            capture_output=True,
            text=True,
            timeout=10,
        )
        agents_md_exists = result.returncode == 0

        # If it exists, check if it contains the COI sandbox context
        if agents_md_exists:
            result = subprocess.run(
                [
                    "incus",
                    "exec",
                    container_name,
                    "--",
                    "grep",
                    "-q",
                    "COI Sandbox Context",
                    "/home/code/.pi/agent/AGENTS.md",
                ],
                capture_output=True,
                text=True,
                timeout=10,
            )
            agents_md_has_context = result.returncode == 0
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

    # Assertions
    assert settings_valid, (
        f"settings.json should be valid JSON. Got: {settings_content}"
    )
    assert agents_md_exists, (
        "~/.pi/agent/AGENTS.md should exist as the auto-context file for pi"
    )
    assert agents_md_has_context, (
        "~/.pi/agent/AGENTS.md should contain COI Sandbox Context"
    )