# Add text as show note for "Science & Technology - Voice of America" podcast

XXX: This is not perfect. It lacks some episodes.

Merge
https://learningenglish.voanews.com/podcast/?zoneId=1579 (podcast)
and
https://learningenglish.voanews.com/api/zmg_pe$myp (text)

## To build
Install pkgsrc/lang/go117 and run:
```
$ GOPATH=~/voa-rss-merger/root go117 build voa-rss-merger.go
```

## To use
Subscribe:
```
http://yourhostname:8080/rss
```
