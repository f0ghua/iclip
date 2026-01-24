"""Configuration management for iclip_server."""

import configparser
from dataclasses import dataclass
from pathlib import Path


@dataclass
class MonitorConfig:
    """Monitor configuration."""
    watch_path: str
    interval: float
    pattern: str


@dataclass
class Config:
    """Main configuration container."""
    monitor: MonitorConfig


def get_config_path() -> Path:
    """Get the path to config.ini file."""
    return Path(__file__).parent / "config.ini"


def load_config(config_path: Path | None = None) -> Config:
    """Load configuration from config.ini file.

    Args:
        config_path: Path to config file. If None, uses default location.

    Returns:
        Config object with loaded settings.

    Raises:
        FileNotFoundError: If config file doesn't exist.
        KeyError: If required config keys are missing.
    """
    if config_path is None:
        config_path = get_config_path()

    if not config_path.exists():
        raise FileNotFoundError(f"Config file not found: {config_path}")

    parser = configparser.ConfigParser()
    parser.read(config_path, encoding="utf-8")

    monitor_config = MonitorConfig(
        watch_path=parser.get("monitor", "watch_path"),
        interval=parser.getfloat("monitor", "interval", fallback=1.0),
        pattern=parser.get("monitor", "pattern", fallback="clipboard_*.tar.gz"),
    )

    return Config(monitor=monitor_config)
