"""Replay stub: the real FHLAPIEndpoints replaced by a router that returns recorded upstream JSON."""
ROUTER = None


class FHLAPIEndpoints:
    async def __aenter__(self):
        return self

    async def __aexit__(self, *a):
        return False

    async def get_audio_bible(self, book_id, chapter, audio_version=0):
        return ROUTER(book_id, chapter, audio_version)
