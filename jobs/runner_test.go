package jobs_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcdole/gofeed"
)

const gazelleFeedTest = `<rss version="2.0">
<channel>
	<title>Good Stuff :: TJSuki 2.0</title>
	<link>http://tjsuki.eu</link>
	<description>Personal RSS feed: Good Stuff</description>
	<language>en-us</language>
	<lastBuildDate>Sat, 28 Mar 2020 17:09:25 +0000</lastBuildDate>
	<docs>http://blogs.law.harvard.edu/tech/rss</docs>
	<generator>Nice Feed Class</generator>
	<item>
		<title>title1</title>
		<description>description1</description>
		<pubDate>Thu, 26 Mar 2020 11:01:04 +0000</pubDate>
		<link>http://tjsuki.eu/torrents.php?action=download&amp;id=01</link>
		<guid>http://tjsuki.eu/torrents.php?action=download&amp;id=01</guid>
		<comments>http://tjsuki.eu/torrents.php?id=01</comments>
	</item>
	<item>
		<title>title2</title>
		<description>description2</description>
		<pubDate>Thu, 26 Mar 2020 22:02:04 +0000</pubDate>
		<link>http://tjsuki.eu/torrents.php?action=download&amp;id=02</link>
		<guid>http://tjsuki.eu/torrents.php?action=download&amp;id=02</guid>
		<comments>http://tjsuki.eu/torrents.php?id=02</comments>
	</item>
</channel>
</rss>`

func TestParseGazelleFeed(t *testing.T) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(gazelleFeedTest)
	if err != nil {
		t.Errorf("coule not parse feed string: %+v", err)
		t.FailNow()
	}
	links := make([]string, len(feed.Items))
	for i, item := range feed.Items {
		links[i] = item.Link
	}
	if diff := cmp.Diff(links, []string{
		"http://tjsuki.eu/torrents.php?action=download&id=01",
		"http://tjsuki.eu/torrents.php?action=download&id=02",
	}); diff != "" {
		t.Error(diff)
	}
}
