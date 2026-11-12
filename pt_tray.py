#!/usr/bin/env python3
"""
BobTrader System Tray & Server Lifecycle Controller (pt_tray.py)

Provides a background monitor and system tray / interactive CLI control
for UltraTrader Go and Python services. Allows checking health status,
viewing real-time statistics, opening the web dashboard in browser,
and cleanly triggering graceful server shutdown via /api/system/shutdown.

Supports graphical system tray mode when pystray + Pillow are available,
and seamlessly provides an interactive CLI/terminal status mode in headless
or server environments.
"""

import sys
import os
import time
import argparse
import webbrowser
import threading
import json
try:
    import urllib.request as urllib_request
    import urllib.error as urllib_error
except ImportError:
    pass

SERVER_DEFAULT_URL = os.environ.get("BOBTRADER_URL", "http://127.0.0.1:8080")

def check_server_status(base_url: str = SERVER_DEFAULT_URL):
    """Checks the health and status of the UltraTrader Go backend."""
    url = f"{base_url}/api/status"
    try:
        req = urllib_request.Request(url, headers={"User-Agent": "BobTrader-Tray/1.0"})
        with urllib_request.urlopen(req, timeout=3) as resp:
            if resp.status == 200:
                data = json.loads(resp.read().decode("utf-8"))
                return True, data
    except Exception as e:
        return False, str(e)
    return False, "Unknown error"

def check_ws_health(base_url: str = SERVER_DEFAULT_URL):
    """Checks the WebSocket feed health."""
    url = f"{base_url}/api/ws-health"
    try:
        req = urllib_request.Request(url, headers={"User-Agent": "BobTrader-Tray/1.0"})
        with urllib_request.urlopen(req, timeout=3) as resp:
            if resp.status == 200:
                data = json.loads(resp.read().decode("utf-8"))
                return data
    except Exception:
        pass
    return None

def trigger_server_shutdown(base_url: str = SERVER_DEFAULT_URL):
    """Sends graceful shutdown POST request to the UltraTrader Go backend."""
    url = f"{base_url}/api/system/shutdown"
    try:
        req = urllib_request.Request(
            url,
            data=b"{}",
            headers={"Content-Type": "application/json", "User-Agent": "BobTrader-Tray/1.0"},
            method="POST"
        )
        with urllib_request.urlopen(req, timeout=5) as resp:
            return resp.status == 200
    except Exception as e:
        print(f"[Tray] Error requesting shutdown: {e}", file=sys.stderr)
        return False

def open_dashboard(base_url: str = SERVER_DEFAULT_URL):
    """Opens the dashboard in the default browser."""
    dashboard_url = f"{base_url}/dashboard"
    print(f"[Tray] Opening dashboard: {dashboard_url}")
    try:
        webbrowser.open(dashboard_url)
    except Exception as e:
        print(f"[Tray] Could not open browser: {e}")

def run_cli_loop(base_url: str = SERVER_DEFAULT_URL, interval: int = 5):
    """Runs a live console monitoring and management session."""
    print("=" * 65)
    print(" BobTrader / UltraTrader Go — System Tray & Process Monitor")
    print(f" Target Server URL: {base_url}")
    print(" Commands: [o] Open Dashboard  [s] Check Status  [q] Shutdown Server & Exit")
    print("=" * 65)

    running = True
    def input_thread():
        nonlocal running
        while running:
            try:
                line = sys.stdin.readline()
                if not line:
                    break
                cmd = line.strip().lower()
                if cmd == 'o':
                    open_dashboard(base_url)
                elif cmd == 's':
                    ok, data = check_server_status(base_url)
                    ws = check_ws_health(base_url)
                    print(f"\n[Status Check] Server Online: {ok} | Details: {data} | WS: {ws}")
                elif cmd == 'q':
                    print("\n[Tray] Initiating graceful shutdown of UltraTrader Go server...")
                    success = trigger_server_shutdown(base_url)
                    if success:
                        print("[Tray] Shutdown command acknowledged by server.")
                    else:
                        print("[Tray] Shutdown command failed or server unreachable.")
                    running = False
                    break
            except Exception:
                break

    t = threading.Thread(target=input_thread, daemon=True)
    t.start()

    while running:
        ok, data = check_server_status(base_url)
        ws = check_ws_health(base_url)
        status_str = "ONLINE (Ready)" if (ok and isinstance(data, dict) and data.get("ready")) else "OFFLINE"
        ws_str = f"Connected ({ws.get('staleness_ms', 0)}ms lag)" if (ws and ws.get("connected")) else "Disconnected"
        sys.stdout.write(f"\r[Tray Monitor] Server: {status_str} | Feed: {ws_str} | {time.strftime('%X')} ")
        sys.stdout.flush()
        time.sleep(interval)
    print("\n[Tray] Monitor terminated.")

def main():
    parser = argparse.ArgumentParser(description="BobTrader System Tray & Server Control")
    parser.add_argument("--url", default=SERVER_DEFAULT_URL, help="Base URL of UltraTrader Go backend")
    parser.add_argument("--interval", type=int, default=5, help="Monitor refresh interval in seconds")
    parser.add_argument("--shutdown", action="store_true", help="Send shutdown signal immediately and exit")
    parser.add_argument("--status", action="store_true", help="Check status immediately and exit")
    args = parser.parse_args()

    if args.shutdown:
        success = trigger_server_shutdown(args.url)
        print(f"Server shutdown triggered: {'Success' if success else 'Failed'}")
        sys.exit(0 if success else 1)

    if args.status:
        ok, data = check_server_status(args.url)
        print(json.dumps({"online": ok, "details": data}, indent=2))
        sys.exit(0 if ok else 1)

    # Check if GUI tray is supported
    try:
        import pystray
        from PIL import Image, ImageDraw
        has_gui = True
    except ImportError:
        has_gui = False

    if has_gui and "DISPLAY" in os.environ:
        print("[Tray] Graphical environment detected, launching GUI tray icon...")
        # Create a simple icon image
        width = 64
        height = 64
        image = Image.new('RGB', (width, height), color=(14, 23, 41))
        dc = ImageDraw.Draw(image)
        dc.rectangle([16, 16, 48, 48], fill=(0, 230, 118))

        def on_open_dashboard(icon, item):
            open_dashboard(args.url)

        def on_shutdown(icon, item):
            trigger_server_shutdown(args.url)
            icon.stop()

        menu = pystray.Menu(
            pystray.MenuItem('Open Dashboard', on_open_dashboard),
            pystray.MenuItem('Shutdown Server & Quit', on_shutdown)
        )
        icon = pystray.Icon("BobTrader", image, "UltraTrader Go", menu)
        icon.run()
    else:
        # Fallback to interactive CLI monitor
        run_cli_loop(base_url=args.url, interval=args.interval)

if __name__ == "__main__":
    main()
