#!/usr/bin/env python3
"""iclip_server - Clipboard transfer server for Windows.

Monitors a shared directory for clipboard archives and sets their
content to the Windows clipboard.
"""

import signal
import sys
import time
from datetime import datetime

try:
    from .clipboard import ClipboardError, set_clipboard_text
    from .config import load_config
    from .monitor import MonitorError, find_clipboard_files, extract_clipboard_data, delete_file
    from .ui import UIState, check_keypress, display_ui
except ImportError:
    from clipboard import ClipboardError, set_clipboard_text
    from config import load_config
    from monitor import MonitorError, find_clipboard_files, extract_clipboard_data, delete_file
    from ui import UIState, check_keypress, display_ui


def main() -> int:
    """Main entry point for iclip_server.

    Returns:
        Exit code (0 for success, non-zero for error).
    """
    # Load configuration
    try:
        config = load_config()
    except FileNotFoundError as e:
        print(f"Error: {e}")
        print("Please create config.ini with your monitor settings.")
        return 1
    except Exception as e:
        print(f"Error loading config: {e}")
        return 1

    # Initialize state
    state = UIState()
    running = True

    def signal_handler(sig, frame):
        nonlocal running
        running = False

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    print("Starting iclip_server...")
    print(f"Monitoring: {config.monitor.watch_path}")
    print(f"Pattern: {config.monitor.pattern}")

    try:
        # Initial display
        display_ui(state)

        while running:
            # Check for keypress
            key = check_keypress()
            if key:
                if key.lower() == "q":
                    running = False
                    continue
                elif key.lower() == "p":
                    state.paused = not state.paused
                    state.status = "Paused" if state.paused else "Monitoring"
                    display_ui(state)

            if state.paused:
                time.sleep(0.1)
                continue

            # Scan for clipboard files
            try:
                files = list(find_clipboard_files(
                    config.monitor.watch_path,
                    config.monitor.pattern,
                ))
            except MonitorError as e:
                state.error_message = str(e)
                state.status = "Error"
                display_ui(state)
                time.sleep(config.monitor.interval)
                continue

            # Process each file
            for file_path in files:
                state.status = "Processing..."
                display_ui(state)

                try:
                    # Extract data from archive
                    data = extract_clipboard_data(file_path)

                    # Set clipboard
                    set_clipboard_text(data.content)

                    # Update state
                    state.received_content = data.content
                    state.received_from = data.username
                    state.last_received = datetime.now()
                    state.status = "Monitoring"
                    state.error_message = ""

                    # Delete processed file
                    delete_file(file_path)

                except (MonitorError, ClipboardError) as e:
                    state.error_message = str(e)
                    state.status = "Error"

                display_ui(state)

            time.sleep(config.monitor.interval)

    except Exception as e:
        print(f"\nError: {e}")
        return 1

    print("\nGoodbye!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
