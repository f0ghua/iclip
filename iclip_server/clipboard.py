"""Clipboard operations module for iclip_server (Windows)."""

import pyperclip


class ClipboardError(Exception):
    """Exception raised for clipboard operation errors."""
    pass


def get_clipboard_text() -> str | None:
    """Get text content from the system clipboard.

    Returns:
        The clipboard text content, or None if clipboard is empty.

    Raises:
        ClipboardError: If clipboard operation fails.
    """
    try:
        content = pyperclip.paste()
        return content if content else None
    except Exception as e:
        raise ClipboardError(f"Failed to get clipboard content: {e}")


def set_clipboard_text(text: str) -> None:
    """Set text content to the system clipboard.

    Args:
        text: The text to set to clipboard.

    Raises:
        ClipboardError: If clipboard operation fails.
    """
    try:
        pyperclip.copy(text)
    except Exception as e:
        raise ClipboardError(f"Failed to set clipboard content: {e}")
