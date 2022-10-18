import time

def handler(req, ctx):
    time.sleep(5)
    return {
        'status_code': 200,
        'content_type': 'application/json',
        'metadata': {
            'key1': 'value1',
            'key2': 'value2'
        },
        'content': {
            'foo': 'bar',
            'baz': 'qux'
        }
    }
