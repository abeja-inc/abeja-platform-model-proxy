################################################################
# for testing proxy programm
################################################################
import asyncio
import importlib
import json
import os
import signal
import socket
import struct
import sys
import tempfile

UNIX_SOCKET_FILE_PATH = '/tmp/samp_v2.sock'
DEFAULT_HANDLER = 'main:handler'

def signal_handler(signum, frame):
    print(f"runtime: signal [{signum}] received")
    loop.stop()

class SocketServer(asyncio.Protocol):
    def __init__(self):
        pass

    def connection_made(self, transport):
        self.transport = transport

    def data_received(self, data):
        print("data:", data)
        req = data[8:]
        req_contents = json.loads(req)
        content = req_contents['contents'][0]
        with open(content['path']) as f:
            s = f.read()
        orgJson = json.loads(s)
        resJson = next(iter(model([orgJson], None)), None)

        b = json.dumps(resJson).encode('utf-8')
        with tempfile.NamedTemporaryFile(delete=False) as fp:
            filename = fp.name
            fp.write(b)

        res = {
            'content_type': 'application/json',
            'path': filename,
            'status_code': 200
        }
        res_str = json.dumps(res).encode('utf-8')
        header = b'\xAB\xE9\xA0\x01' + struct.pack('>I', len(res_str))
        self.transport.write(header)
        self.transport.write(res_str)

    def connection_lost(self, exc):
        self.transport.close()

def import_model(handler: str):
    module_name, func_name = handler.split(':', 1)
    sys.path.insert(0, '.')
    model = getattr(importlib.import_module(module_name), func_name)
    sys.path.remove('.')
    return model

# main
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

handler = os.environ.get('HANDLER', DEFAULT_HANDLER)
model = import_model(handler)

loop = asyncio.get_event_loop()
server_coroutine = loop.create_unix_server(SocketServer, UNIX_SOCKET_FILE_PATH)
server = loop.run_until_complete(server_coroutine)

try:
    loop.run_forever()
except Exception as e:
    print(f"unexpected error: {e}")
finally:
    server.close()
    loop.close()
    os.remove(UNIX_SOCKET_FILE_PATH)

sys.exit(0)
