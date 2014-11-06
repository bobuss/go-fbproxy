import asyncio
import aiohttp

urls = ['http://localhost:3000/proxy/%s' % i for i in range(4, 115)]


@asyncio.coroutine
def get(*args, **kwargs):
    response = yield from aiohttp.request('GET', *args, **kwargs)
    return (yield from response.read_and_close(decode=False))


@asyncio.coroutine
def scrap(url):
    page = yield from get(url, compress=True)
    print(page)
    return page

loop = asyncio.get_event_loop()
f = asyncio.wait([scrap(url) for url in urls])
loop.run_until_complete(f)