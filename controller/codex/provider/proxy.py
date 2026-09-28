#!/usr/bin/env python3
"""A CONNECT-only egress gate for a single sandboxed provider.

The proxy does not terminate TLS. Every attempted destination is journaled
before a connection is opened or refused. Only exact host names and TCP 443
are accepted; IP literals and ordinary HTTP requests are refused.
"""
import argparse
import ipaddress
import json
import os
from pathlib import Path
import selectors
import socket
import socketserver
import threading
import time


def append_observation(path, observation):
    line = json.dumps(observation, sort_keys=True, separators=(',', ':')).encode() + b'\n'
    with open(path, 'ab', buffering=0) as journal:
        journal.write(line)
        os.fsync(journal.fileno())


def destination(authority):
    if authority.count(':') != 1:
        return None
    host, port = authority.rsplit(':', 1)
    if port != '443' or not host or host.lower() != host or host.endswith('.'):
        return None
    try:
        ipaddress.ip_address(host)
        return None
    except ValueError:
        pass
    if any(not label or not all(c.isascii() and (c.isalnum() or c == '-') for c in label)
           for label in host.split('.')):
        return None
    return host


class Gate(socketserver.ThreadingMixIn, socketserver.TCPServer):
    daemon_threads = True
    allow_reuse_address = False

    def __init__(self, address, allow, journal):
        self.allow = frozenset(allow)
        self.journal = journal
        self.lock = threading.Lock()
        super().__init__(address, Tunnel)

    def observe(self, host, decision):
        with self.lock:
            append_observation(self.journal, {'at_ns': time.time_ns(), 'host': host, 'decision': decision})


class Tunnel(socketserver.StreamRequestHandler):
    rbufsize = 0  # do not prefetch TLS bytes after CONNECT headers

    def handle(self):
        self.connection.settimeout(15)
        request = self.rfile.readline(8193)
        if len(request) > 8192 or not request.endswith(b'\r\n'):
            return
        try:
            method, authority, version = request.decode('ascii').strip().split(' ')
        except (UnicodeDecodeError, ValueError):
            return
        headers = bytearray()
        while len(headers) < 32768:
            line = self.rfile.readline(8193)
            if not line or len(line) > 8192:
                return
            headers.extend(line)
            if line == b'\r\n':
                break
        else:
            return
        host = destination(authority) if method == 'CONNECT' and version == 'HTTP/1.1' else None
        if host is None or host not in self.server.allow:
            self.server.observe(authority, 'refused')
            self.wfile.write(b'HTTP/1.1 403 Forbidden\r\nContent-Length: 0\r\n\r\n')
            return
        self.server.observe(host, 'allowed')
        try:
            upstream = socket.create_connection((host, 443), timeout=15)
        except OSError:
            self.wfile.write(b'HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n')
            return
        with upstream:
            self.wfile.write(b'HTTP/1.1 200 Connection Established\r\n\r\n')
            self.wfile.flush()
            self.connection.settimeout(None)
            upstream.settimeout(None)
            with selectors.DefaultSelector() as selector:
                selector.register(self.connection, selectors.EVENT_READ, upstream)
                selector.register(upstream, selectors.EVENT_READ, self.connection)
                while selector.get_map():
                    for key, _ in selector.select():
                        source, target = key.fileobj, key.data
                        try:
                            chunk = source.recv(65536)
                            if chunk:
                                target.sendall(chunk)
                            else:
                                selector.unregister(source)
                                target.shutdown(socket.SHUT_WR)
                        except OSError:
                            return


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--allow', action='append', required=True)
    parser.add_argument('--journal', type=Path, required=True)
    parser.add_argument('--ready', type=Path, required=True)
    parser.add_argument('--watch-fd', type=int)
    args = parser.parse_args()
    if any(destination(host + ':443') != host for host in args.allow):
        parser.error('allow entries must be lower-case DNS host names')
    with Gate(('127.0.0.1', 0), args.allow, args.journal) as gate:
        if args.watch_fd is not None:
            def watch():
                try:
                    while os.read(args.watch_fd, 4096):
                        pass
                finally:
                    gate.shutdown()
            threading.Thread(target=watch, daemon=True).start()
        args.ready.write_text(str(gate.server_address[1]))
        gate.serve_forever(poll_interval=0.25)


if __name__ == '__main__':
    main()
