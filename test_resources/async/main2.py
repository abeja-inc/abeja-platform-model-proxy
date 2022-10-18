import os
import sys
import time

def handler(req, ctx):
    st = os.environ.get("SLEEP_TIME", "5")
    time.sleep(int(st))
    sys.exit(128)
