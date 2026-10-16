"""Replay stub: the real FHLAPIEndpoints replaced by a router that returns recorded upstream JSON."""
ROUTER = None


class FHLAPIEndpoints:
    async def __aenter__(self):
        return self

    async def __aexit__(self, *a):
        return False

    async def get_commentary(self, book, chapter, verse, commentary_id=None):
        return ROUTER("get_commentary", book=book, chapter=chapter, verse=verse)

    async def list_commentaries(self):
        return ROUTER("list_commentaries")

    async def search_commentary(self, keyword, commentary_id=None):
        return ROUTER("search_commentary", keyword=keyword)

    async def get_topic_study(self, keyword=None, topic_id=None, source="all", count_only=False):
        return ROUTER("get_topic_study", keyword=keyword, source=source, count_only=count_only)
