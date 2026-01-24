"""Terminal UI module for iclip_server."""

import os
import sys
import msvcrt
from dataclasses import dataclass
from datetime import datetime
from typing import Optional


@dataclass
class UIState:
    """State for the UI display."""
    status: str = "Monitoring"
    last_received: Optional[datetime] = None
    received_content: str = ""
    received_from: str = ""
    error_message: str = ""
    paused: bool = False


VERSION = "1.0"
MAX_CONTENT_LINES = 5
MAX_LINE_WIDTH = 34


def truncate_text(text: str, max_lines: int = MAX_CONTENT_LINES, max_width: int = MAX_LINE_WIDTH) -> list[str]:
    """Truncate text to fit in the UI box.

    Args:
        text: The text to truncate.
        max_lines: Maximum number of lines to show.
        max_width: Maximum width of each line.

    Returns:
        List of lines to display.
    """
    if not text:
        return ["(empty)"]

    lines = text.split("\n")
    result = []

    for line in lines:
        if len(result) >= max_lines:
            break
        if len(line) > max_width:
            result.append(line[:max_width - 3] + "...")
        else:
            result.append(line)

    if len(lines) > max_lines:
        if result:
            result[-1] = "..."

    return result if result else ["(empty)"]


def clear_screen() -> None:
    """Clear the terminal screen."""
    os.system("cls" if os.name == "nt" else "clear")


def render_ui(state: UIState) -> str:
    """Render the UI to a string.

    Args:
        state: Current UI state.

    Returns:
        Rendered UI string.
    """
    status_indicator = "●" if not state.paused else "○"
    status_text = "Paused" if state.paused else state.status

    last_received_str = state.last_received.strftime("%H:%M:%S") if state.last_received else "Never"

    content_lines = truncate_text(state.received_content)

    lines = [
        "╔══════════════════════════════════════╗",
        "║         iclip_server v{}            ║".format(VERSION),
        "╠══════════════════════════════════════╣",
        "║ Status: {} {:27}║".format(status_indicator, status_text),
        "║ Last received: {:22}║".format(last_received_str),
    ]

    if state.received_from:
        lines.append("║ From: {:31}║".format(state.received_from[:31]))

    lines.extend([
        "╠══════════════════════════════════════╣",
        "║ Received content:                    ║",
        "║ ┌────────────────────────────────┐   ║",
    ])

    for content_line in content_lines:
        padded = content_line.ljust(MAX_LINE_WIDTH)
        lines.append("║ │ {} │   ║".format(padded[:MAX_LINE_WIDTH]))

    # Fill remaining lines if needed
    for _ in range(MAX_CONTENT_LINES - len(content_lines)):
        lines.append("║ │ {:34} │   ║".format(""))

    lines.extend([
        "║ └────────────────────────────────┘   ║",
        "╠══════════════════════════════════════╣",
    ])

    if state.error_message:
        error_display = state.error_message[:34]
        lines.append("║ Error: {:30}║".format(error_display))
        lines.append("╠══════════════════════════════════════╣")

    lines.extend([
        "║ [q] Quit  [p] Pause                  ║",
        "╚══════════════════════════════════════╝",
    ])

    return "\n".join(lines)


def display_ui(state: UIState) -> None:
    """Display the UI on the terminal.

    Args:
        state: Current UI state.
    """
    clear_screen()
    print(render_ui(state))


def check_keypress() -> Optional[str]:
    """Check for a keypress without blocking (Windows).

    Returns:
        The key pressed, or None if no key was pressed.
    """
    try:
        if msvcrt.kbhit():
            return msvcrt.getch().decode("utf-8", errors="replace")
    except Exception:
        pass
    return None
