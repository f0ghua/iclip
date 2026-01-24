"""Directory monitoring and file processing module for iclip_server."""

import fnmatch
import json
import os
import shutil
import tarfile
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Generator


class MonitorError(Exception):
    """Exception raised for monitor operation errors."""
    pass


@dataclass
class ClipboardData:
    """Parsed clipboard data from archive."""
    content: str
    content_type: str
    username: str
    timestamp: str


def find_clipboard_files(watch_path: str, pattern: str) -> Generator[Path, None, None]:
    """Find clipboard archive files matching the pattern.

    Args:
        watch_path: Directory path to watch.
        pattern: Glob pattern for matching files (e.g., "clipboard_*.tar.gz").

    Yields:
        Path objects for matching files.

    Raises:
        MonitorError: If the watch directory doesn't exist or is inaccessible.
    """
    try:
        watch_dir = Path(watch_path)
        if not watch_dir.exists():
            raise MonitorError(f"Watch directory does not exist: {watch_path}")

        for item in watch_dir.iterdir():
            if item.is_file() and fnmatch.fnmatch(item.name, pattern):
                yield item
    except PermissionError:
        raise MonitorError(f"Permission denied accessing: {watch_path}")
    except Exception as e:
        if isinstance(e, MonitorError):
            raise
        raise MonitorError(f"Error scanning directory: {e}")


def extract_clipboard_data(archive_path: Path) -> ClipboardData:
    """Extract and parse clipboard data from a tar.gz archive.

    Args:
        archive_path: Path to the clipboard archive file.

    Returns:
        ClipboardData object with parsed content.

    Raises:
        MonitorError: If extraction or parsing fails.
    """
    temp_dir = None
    try:
        temp_dir = tempfile.mkdtemp(prefix="iclip_extract_")

        # Extract archive
        with tarfile.open(archive_path, "r:gz") as tar:
            # Security check: prevent path traversal
            for member in tar.getmembers():
                if os.path.isabs(member.name) or ".." in member.name:
                    raise MonitorError("Archive contains unsafe paths")
            tar.extractall(temp_dir)

        # Find the clipboard_data directory
        data_dir = Path(temp_dir) / "clipboard_data"
        if not data_dir.exists():
            raise MonitorError("Archive missing clipboard_data directory")

        # Read meta.json
        meta_file = data_dir / "meta.json"
        if not meta_file.exists():
            raise MonitorError("Archive missing meta.json")

        with open(meta_file, "r", encoding="utf-8") as f:
            meta = json.load(f)

        content_type = meta.get("type", "text")
        username = meta.get("username", "unknown")
        timestamp = meta.get("timestamp", "")

        # Read content based on type
        if content_type == "text":
            content_file = data_dir / "content.txt"
            if not content_file.exists():
                raise MonitorError("Archive missing content.txt")
            content = content_file.read_text(encoding="utf-8")
        else:
            raise MonitorError(f"Unsupported content type: {content_type}")

        return ClipboardData(
            content=content,
            content_type=content_type,
            username=username,
            timestamp=timestamp,
        )

    except tarfile.TarError as e:
        raise MonitorError(f"Failed to extract archive: {e}")
    except json.JSONDecodeError as e:
        raise MonitorError(f"Failed to parse meta.json: {e}")
    except MonitorError:
        raise
    except Exception as e:
        raise MonitorError(f"Error processing archive: {e}")
    finally:
        # Cleanup temp directory
        if temp_dir:
            shutil.rmtree(temp_dir, ignore_errors=True)


def delete_file(file_path: Path) -> None:
    """Delete a file safely.

    Args:
        file_path: Path to the file to delete.

    Raises:
        MonitorError: If deletion fails.
    """
    try:
        file_path.unlink()
    except PermissionError:
        raise MonitorError(f"Permission denied deleting: {file_path}")
    except Exception as e:
        raise MonitorError(f"Failed to delete file: {e}")
